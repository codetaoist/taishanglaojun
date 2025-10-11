package models

import (
	"time"
	"gorm.io/gorm"
)

// User 简单用户模型，用于聊天系统
type User struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Username string `json:"username" gorm:"size:50;not null"`
	Nickname string `json:"nickname" gorm:"size:100"`
}

// ChatRoom 聊天室模�?
type ChatRoom struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"size:100;not null"`
	Description string         `json:"description" gorm:"size:500"`
	Type        string         `json:"type" gorm:"size:20;not null;default:'public'"` // public, private, group
	CreatorID   uint           `json:"creator_id" gorm:"not null"`
	MaxMembers  int            `json:"max_members" gorm:"default:100"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// 关联
	Creator  User              `json:"creator,omitempty" gorm:"foreignKey:CreatorID"`
	Members  []ChatRoomMember  `json:"members,omitempty" gorm:"foreignKey:RoomID"`
	Messages []ChatMessage     `json:"messages,omitempty" gorm:"foreignKey:RoomID"`
}

// ChatRoomMember 聊天室成员模�?
type ChatRoomMember struct {
	ID       uint      `json:"id" gorm:"primaryKey"`
	RoomID   uint      `json:"room_id" gorm:"not null"`
	UserID   uint      `json:"user_id" gorm:"not null"`
	Role     string    `json:"role" gorm:"size:20;default:'member'"` // admin, moderator, member
	JoinedAt time.Time `json:"joined_at" gorm:"default:CURRENT_TIMESTAMP"`
	IsActive bool      `json:"is_active" gorm:"default:true"`

	// 关联
	Room User `json:"room,omitempty" gorm:"foreignKey:RoomID"`
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`

	// 唯一索引
	// gorm:"uniqueIndex:idx_room_user"
}

// ChatMessage 聊天消息模型
type ChatMessage struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	RoomID    uint           `json:"room_id" gorm:"not null"`
	SenderID  uint           `json:"sender_id" gorm:"not null"`
	Content   string         `json:"content" gorm:"type:text;not null"`
	Type      string         `json:"type" gorm:"size:20;default:'text'"` // text, image, file, system
	ReplyToID *uint          `json:"reply_to_id,omitempty"`              // 回复的消息ID
	IsEdited  bool           `json:"is_edited" gorm:"default:false"`
	IsDeleted bool           `json:"is_deleted" gorm:"default:false"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// 关联
	Room     ChatRoom     `json:"room,omitempty" gorm:"foreignKey:RoomID"`
	Sender   User         `json:"sender,omitempty" gorm:"foreignKey:SenderID"`
	ReplyTo  *ChatMessage `json:"reply_to,omitempty" gorm:"foreignKey:ReplyToID"`
	Replies  []ChatMessage `json:"replies,omitempty" gorm:"foreignKey:ReplyToID"`
}

// ChatMessageRead 消息读取状态模�?
type ChatMessageRead struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	MessageID uint      `json:"message_id" gorm:"not null"`
	UserID    uint      `json:"user_id" gorm:"not null"`
	ReadAt    time.Time `json:"read_at" gorm:"default:CURRENT_TIMESTAMP"`

	// 关联
	Message ChatMessage `json:"message,omitempty" gorm:"foreignKey:MessageID"`
	User    User        `json:"user,omitempty" gorm:"foreignKey:UserID"`

	// 唯一索引
	// gorm:"uniqueIndex:idx_message_user"
}

// PrivateChat 私聊模型
type PrivateChat struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	User1ID     uint           `json:"user1_id" gorm:"not null"`
	User2ID     uint           `json:"user2_id" gorm:"not null"`
	LastMessage string         `json:"last_message" gorm:"type:text"`
	LastMsgTime *time.Time     `json:"last_msg_time"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// 关联
	User1    User                `json:"user1,omitempty" gorm:"foreignKey:User1ID"`
	User2    User                `json:"user2,omitempty" gorm:"foreignKey:User2ID"`
	Messages []PrivateChatMessage `json:"messages,omitempty" gorm:"foreignKey:ChatID"`

	// 唯一索引确保两个用户之间只有一个私聊记�?
	// gorm:"uniqueIndex:idx_user1_user2"
}

// PrivateChatMessage 私聊消息模型
type PrivateChatMessage struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	ChatID    uint           `json:"chat_id" gorm:"not null"`
	SenderID  uint           `json:"sender_id" gorm:"not null"`
	Content   string         `json:"content" gorm:"type:text;not null"`
	Type      string         `json:"type" gorm:"size:20;default:'text'"` // text, image, file
	IsRead    bool           `json:"is_read" gorm:"default:false"`
	IsDeleted bool           `json:"is_deleted" gorm:"default:false"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// 关联
	Chat   PrivateChat `json:"chat,omitempty" gorm:"foreignKey:ChatID"`
	Sender User        `json:"sender,omitempty" gorm:"foreignKey:SenderID"`
}

// OnlineUser 在线用户模型（用于缓存）
type OnlineUser struct {
	UserID       uint      `json:"user_id" gorm:"primaryKey"`
	LastSeen     time.Time `json:"last_seen" gorm:"default:CURRENT_TIMESTAMP"`
	Status       string    `json:"status" gorm:"size:20;default:'online'"` // online, away, busy, offline
	SocketID     string    `json:"socket_id" gorm:"size:100"`
	IPAddress    string    `json:"ip_address" gorm:"size:45"`
	UserAgent    string    `json:"user_agent" gorm:"size:500"`
	ConnectedAt  time.Time `json:"connected_at" gorm:"default:CURRENT_TIMESTAMP"`

	// 关联
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// ChatNotification 聊天通知模型
type ChatNotification struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    uint           `json:"user_id" gorm:"not null"`
	Type      string         `json:"type" gorm:"size:50;not null"`        // message, mention, room_invite
	Title     string         `json:"title" gorm:"size:200;not null"`
	Content   string         `json:"content" gorm:"type:text"`
	Data      string         `json:"data" gorm:"type:json"`               // 额外数据，JSON格式
	IsRead    bool           `json:"is_read" gorm:"default:false"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// 关联
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName 设置表名
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
