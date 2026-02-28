package model

import "time"

type ProjectStatus string

const (
	ProjectActive   ProjectStatus = "active"
	ProjectArchived ProjectStatus = "archived"
)

type TaskStatus string

const (
	TaskTodo       TaskStatus = "todo"
	TaskInProgress TaskStatus = "in_progress"
	TaskDone       TaskStatus = "done"
)

type Project struct {
	ID          uint          `json:"id" gorm:"primaryKey"`
	Title       string        `json:"title" gorm:"not null;index"`
	Description string        `json:"description"`
	Status      ProjectStatus `json:"status" gorm:"not null;index"`
	CreatedAt   time.Time     `json:"createdAt" gorm:"index"`
	UpdatedAt   time.Time     `json:"updatedAt"`

	Tasks []Task `json:"tasks,omitempty" gorm:"constraint:OnDelete:CASCADE;"`
}

type Task struct {
	ID        uint       `json:"id" gorm:"primaryKey"`
	ProjectID uint       `json:"projectId" gorm:"not null;index"`
	Title     string     `json:"title" gorm:"not null;index"`
	Status    TaskStatus `json:"status" gorm:"not null;index"`
	Assignee  string     `json:"assignee" gorm:"index"`
	DueDate   *time.Time `json:"dueDate,omitempty" gorm:"index"`
	CreatedAt time.Time  `json:"createdAt" gorm:"index"`
	UpdatedAt time.Time  `json:"updatedAt"`

	Comments []Comment `json:"comments,omitempty" gorm:"constraint:OnDelete:CASCADE;"`
}

type Comment struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	TaskID    uint      `json:"taskId" gorm:"not null;index"`
	Author    string    `json:"author" gorm:"not null;index"`
	Text      string    `json:"text" gorm:"not null"`
	CreatedAt time.Time `json:"createdAt" gorm:"index"`
	UpdatedAt time.Time `json:"updatedAt"`
}
