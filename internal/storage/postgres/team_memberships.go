package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robwittman/pillar/internal/domain"
)

type TeamMembershipRepository struct {
	pool *pgxpool.Pool
}

func NewTeamMembershipRepository(pool *pgxpool.Pool) *TeamMembershipRepository {
	return &TeamMembershipRepository{pool: pool}
}

func (r *TeamMembershipRepository) Add(ctx context.Context, tm *domain.TeamMembership) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO team_memberships (id, team_id, user_id)
		 VALUES ($1, $2, $3)
		 RETURNING created_at`,
		tm.ID, tm.TeamID, tm.UserID,
	).Scan(&tm.CreatedAt)
	if err != nil {
		return fmt.Errorf("inserting team membership: %w", err)
	}
	return nil
}

func (r *TeamMembershipRepository) Remove(ctx context.Context, teamID, userID string) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM team_memberships WHERE team_id = $1 AND user_id = $2`,
		teamID, userID,
	)
	if err != nil {
		return fmt.Errorf("deleting team membership: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrMembershipNotFound
	}
	return nil
}

func (r *TeamMembershipRepository) ListByTeam(ctx context.Context, teamID string) ([]*domain.TeamMembership, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, team_id, user_id, created_at
		 FROM team_memberships WHERE team_id = $1 ORDER BY created_at ASC`, teamID)
	if err != nil {
		return nil, fmt.Errorf("querying team memberships: %w", err)
	}
	defer rows.Close()

	var memberships []*domain.TeamMembership
	for rows.Next() {
		tm := &domain.TeamMembership{}
		if err := rows.Scan(&tm.ID, &tm.TeamID, &tm.UserID, &tm.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning team membership: %w", err)
		}
		memberships = append(memberships, tm)
	}
	return memberships, rows.Err()
}

func (r *TeamMembershipRepository) ListByUser(ctx context.Context, userID string) ([]*domain.TeamMembership, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, team_id, user_id, created_at
		 FROM team_memberships WHERE user_id = $1 ORDER BY created_at ASC`, userID)
	if err != nil {
		return nil, fmt.Errorf("querying team memberships: %w", err)
	}
	defer rows.Close()

	var memberships []*domain.TeamMembership
	for rows.Next() {
		tm := &domain.TeamMembership{}
		if err := rows.Scan(&tm.ID, &tm.TeamID, &tm.UserID, &tm.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning team membership: %w", err)
		}
		memberships = append(memberships, tm)
	}
	return memberships, rows.Err()
}
