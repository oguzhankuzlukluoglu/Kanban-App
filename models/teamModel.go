package models

import (
	"gorm.io/gorm"
)

type Team struct {
	ID      uint   `json:"id" gorm:"primaryKey"`
	Name    string `json:"name"`
	Members []TeamMember
}

type TeamMember struct {
	gorm.Model
	TeamID      uint
	UserID      uint
	InvitedByID uint
	User        TeamUser
	InvitedBy   InvitedBy `gorm:"foreignKey:InvitedByID"`
}

type TeamUser struct {
	ID         uint   `json:"id" gorm:"primaryKey"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	Role       int    `json:"role"`
	CustomRole string `json:"custom_role"`
}

type Seats struct {
	Members Members `json:"members"`
}

type Members struct {
	FilledMembersSeats int `json:"filled_members_seats"`
	TotalMemberSeats   int `json:"total_member_seats"`
	EmptyMemberSeats   int `json:"empty_member_seats"`
}

type InvitedBy struct {
	ID             uint   `json:"id" gorm:"primaryKey"`
	Username       string `json:"username"`
	Email          string `json:"email"`
	ProfilePicture string `json:"profilePicture"`
}