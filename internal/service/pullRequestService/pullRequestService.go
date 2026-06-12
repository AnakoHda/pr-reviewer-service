package pullRequestService

import (
	"pr-reviewer-service/internal/domain"
	"context"
)

type UsersRepository interface {
	GetUserByID(ctx context.Context, userID domain.UserId) (*domain.User, error)
	ListUsersByTeamName(ctx context.Context, teamName string) ([]domain.User, error)
}
type PRRepository interface {
	GetPullRequestByID(ctx context.Context, pullRequestId domain.PullRequestId) (*domain.PullRequest, error)
	CreatePullRequest(ctx context.Context, pullRequest domain.PullRequest) error
	UpdatePullRequest(ctx context.Context, pullRequest domain.PullRequest) error
}
type Service struct {
	userRepo UsersRepository
	prRepo   PRRepository
}

func New(usersRepository UsersRepository, prRepository PRRepository) *Service {
	return &Service{
		userRepo: usersRepository,
		prRepo:   prRepository,
	}
}

func (s *Service) CreatePullRequest(ctx context.Context, PullRequestId domain.PullRequestId, PullRequestName string, AuthorId domain.UserId) (*domain.PullRequest, error) {
	//проверка что автор/команда существует
	author, err := s.userRepo.GetUserByID(ctx, AuthorId)
	if err != nil {
		return nil, err
	}
	//получение []User состоящих в команде автора
	usersByTeam, err := s.userRepo.ListUsersByTeamName(ctx, author.TeamName)
	if err != nil {
		return nil, err
	}
	//нахожденте подходящих кондидатов
	candidates := make([]domain.UserId, 0, 2)
	for _, user := range usersByTeam {
		if user.IsActive && user.UserId != AuthorId {
			candidates = append(candidates, user.UserId)
		}
		if len(candidates) == 2 {
			break
		}
	}
	//создание нового PR
	newPullRequest, err := domain.NewPullRequest(PullRequestId, PullRequestName, AuthorId, candidates)
	if err != nil {
		return nil, err
	}
	//сохранение PR в репозитории
	if err := s.prRepo.CreatePullRequest(ctx, *newPullRequest); err != nil {
		return nil, err
	}

	return newPullRequest, nil

}

func (s *Service) Merge(ctx context.Context, pullRequestID domain.PullRequestId) (*domain.PullRequest, error) {
	//провека наличия PR
	foundedPR, err := s.prRepo.GetPullRequestByID(ctx, pullRequestID)
	if err != nil {
		return nil, domain.ErrPullRequestNotFound
	}
	//Проверка не является ли он уже в статусе MERGED
	if foundedPR.Merge() {
		//обновление статуса в базе данных
		if err := s.prRepo.UpdatePullRequest(ctx, *foundedPR); err != nil {
			return nil, err
		}
	}
	return foundedPR, nil
}

func (s *Service) Reassign(ctx context.Context, pullRequestID domain.PullRequestId, oldReviewerId domain.UserId) (*domain.PullRequest, *domain.UserId, error) {
	//поиск PR
	foundedPR, err := s.prRepo.GetPullRequestByID(ctx, pullRequestID)
	if err != nil {
		return nil, nil, domain.ErrPullRequestNotFound
	}
	//поверка статуса, oldReviewerId != AuthorId
	if foundedPR.AuthorId == oldReviewerId {
		return nil, nil, domain.ErrAuthorCannotBeReviewer
	}
	if foundedPR.Status == domain.PullRequestStatusMERGED {
		return nil, nil, domain.ErrPullRequestMerged
	}
	//поиск старого User
	oldUserReviewer, err := s.userRepo.GetUserByID(ctx, oldReviewerId)
	if err != nil {
		return nil, nil, domain.ErrUserNotFound
	}
	//ищем в команде oldUserReviewer потенциального рревьювера
	usersByTeam, err := s.userRepo.ListUsersByTeamName(ctx, oldUserReviewer.TeamName)
	if err != nil {
		return nil, nil, err
	}

	for _, user := range usersByTeam {
		if !user.IsActive || user.UserId == oldUserReviewer.UserId {
			continue
		}
		if err := foundedPR.ReplaceReviewer(oldReviewerId, user.UserId); err != nil {
			if err.Error() == domain.ErrNotFoundReviewerInPullRequest.Error() {
				return nil, nil, domain.ErrNotFoundReviewerInPullRequest
			}
			continue
		}
		if err := s.prRepo.UpdatePullRequest(ctx, *foundedPR); err != nil {
			return nil, nil, err
		}

		return foundedPR, &user.UserId, nil
	}
	return nil, nil, domain.ErrNoCandidatesInTeam
}
