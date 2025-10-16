package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User 
type User struct {
	ID       uuid.UUID `json:"id" gorm:"type:char(36);primaryKey"`
	Username string    `json:"username" gorm:"size:50;not null"`
	Nickname string    `json:"nickname" gorm:"size:100"`
}

// ChatRoom 
type ChatRoom struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"size:100;not null"`
	Description string         `json:"description" gorm:"size:500"`
	Type        string         `json:"type" gorm:"size:20;not null;default:'public'"` // public, private, group
	CreatorID   uuid.UUID      `json:"creator_id" gorm:"type:char(36);not null"`
	MaxMembers  int            `json:"max_members" gorm:"default:100"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// 
	Creator  User             `json:"creator,omitempty" gorm:"foreignKey:CreatorID"`
	Members  []ChatRoomMember `json:"members,omitempty" gorm:"foreignKey:RoomID"`
	Messages []ChatMessage    `json:"messages,omitempty" gorm:"foreignKey:RoomID"`
}

// ChatRoomMember 
type ChatRoomMember struct {
	ID       uint      `json:"id" gorm:"primaryKey"`
	RoomID   uint      `json:"room_id" gorm:"not null"`
	UserID   uuid.UUID `json:"user_id" gorm:"type:char(36);not null"`
	Role     string    `json:"role" gorm:"size:20;default:'member'"` // admin, moderator, member
	JoinedAt time.Time `json:"joined_at" gorm:"default:CURRENT_TIMESTAMP"`
	IsActive bool      `json:"is_active" gorm:"default:true"`

	// 
	Room ChatRoom `json:"room,omitempty" gorm:"foreignKey:RoomID"`
	User User     `json:"user,omitempty" gorm:"foreignKey:UserID"`

	// 
	// gorm:"uniqueIndex:idx_room_user"
}

// ChatMessage 
type ChatMessage struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	RoomID    uint           `json:"room_id" gorm:"not null"`
	SenderID  uuid.UUID      `json:"sender_id" gorm:"type:char(36);not null"`
	Content   string         `json:"content" gorm:"type:text;not null"`
	Type      string         `json:"type" gorm:"size:20;default:'text'"` // text, image, file, system
	ReplyToID *uint          `json:"reply_to_id,omitempty"`              // ID
	IsEdited  bool           `json:"is_edited" gorm:"default:false"`
	IsDeleted bool           `json:"is_deleted" gorm:"default:false"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// 
	Room    ChatRoom      `json:"room,omitempty" gorm:"foreignKey:RoomID"`
	Sender  User          `json:"sender,omitempty" gorm:"foreignKey:SenderID"`
	ReplyTo *ChatMessage  `json:"reply_to,omitempty" gorm:"foreignKey:ReplyToID"`
	Replies []ChatMessage `json:"replies,omitempty" gorm:"foreignKey:ReplyToID"`
}

// ChatMessageRead 
type ChatMessageRead struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	MessageID uint      `json:"message_id" gorm:"not null"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:char(36);not null"`
	ReadAt    time.Time `json:"read_at" gorm:"default:CURRENT_TIMESTAMP"`

	// 
	Message ChatMessage `json:"message,omitempty" gorm:"foreignKey:MessageID"`
	User    User        `json:"user,omitempty" gorm:"foreignKey:UserID"`

	// 
	// gorm:"uniqueIndex:idx_message_user"
}

// PrivateChat 
type PrivateChat struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	User1ID     uuid.UUID      `json:"user1_id" gorm:"type:char(36);not null"`
	User2ID     uuid.UUID      `json:"user2_id" gorm:"type:char(36);not null"`
	LastMessage string         `json:"last_message" gorm:"type:text"`
	LastMsgTime *time.Time     `json:"last_msg_time"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// 
	User1    User                 `json:"user1,omitempty" gorm:"foreignKey:User1ID"`
	User2    User                 `json:"user2,omitempty" gorm:"foreignKey:User2ID"`
	Messages []PrivateChatMessage `json:"messages,omitempty" gorm:"foreignKey:ChatID"`

	// 
	// gorm:"uniqueIndex:idx_user1_user2"
}

// PrivateChatMessage 
type PrivateChatMessage struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	ChatID    uint           `json:"chat_id" gorm:"not null"`
	SenderID  uuid.UUID      `json:"sender_id" gorm:"type:char(36);not null"`
	Content   string         `json:"content" gorm:"type:text;not null"`
	Type      string         `json:"type" gorm:"size:20;default:'text'"` // text, image, file
	IsRead    bool           `json:"is_read" gorm:"default:false"`
	IsDeleted bool           `json:"is_deleted" gorm:"default:false"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// 
	Chat   PrivateChat `json:"chat,omitempty" gorm:"foreignKey:ChatID"`
	Sender User        `json:"sender,omitempty" gorm:"foreignKey:SenderID"`
}

// OnlineUser 
type OnlineUser struct {
	UserID      uuid.UUID `json:"user_id" gorm:"type:char(36);primaryKey"`
	LastSeen    time.Time `json:"last_seen" gorm:"default:CURRENT_TIMESTAMP"`
	Status      string    `json:"status" gorm:"size:20;default:'online'"` // online, away, busy, offline
	SocketID    string    `json:"socket_id" gorm:"size:100"`
	IPAddress   string    `json:"ip_address" gorm:"size:45"`
	UserAgent   string    `json:"user_agent" gorm:"size:500"`
	ConnectedAt time.Time `json:"connected_at" gorm:"default:CURRENT_TIMESTAMP"`

	// 
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// ChatNotification 
type ChatNotification struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    uuid.UUID      `json:"user_id" gorm:"type:char(36);not null"`
	Type      string         `json:"type" gorm:"size:50;not null"` // message, mention, room_invite
	Title     string         `json:"title" gorm:"size:200;not null"`
	Content   string         `json:"content" gorm:"type:text"`
	Data      string         `json:"data" gorm:"type:json"` // JSON
	IsRead    bool           `json:"is_read" gorm:"default:false"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// 
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName 
func (ChatRoom) TableName() string {
	return "chat_rooms"
}

func (ChatRoomMember) TableName() string {
	return "chat_room_members"
}

func (ChatMessage) TableName() string {
	return "chat_messages"
}

func (ChatMessageRead) TableName() string {
	return "chat_message_reads"
}

func (PrivateChat) TableName() string {
	return "private_chats"
}

func (PrivateChatMessage) TableName() string {
	return "private_chat_messages"
}

func (OnlineUser) TableName() string {
	return "online_users"
}

func (ChatNotification) TableName() string {
	return "chat_notifications"
}

