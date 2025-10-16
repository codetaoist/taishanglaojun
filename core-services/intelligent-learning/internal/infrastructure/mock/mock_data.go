package mock

import (
	"time"

	"github.com/google/uuid"
)

// MockLearnerProfile ?
type MockLearnerProfile struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Level        int       `json:"level"`
	Experience   int       `json:"experience"`
	Avatar       string    `json:"avatar"`
	JoinDate     time.Time `json:"joinDate"`
	StudyStreak  int       `json:"studyStreak"`
	TotalStudyTime int     `json:"totalStudyTime"`
}

// MockLearningAnalytics 
type MockLearningAnalytics struct {
	TotalStudyTime    int     `json:"totalStudyTime"`
	CompletedCourses  int     `json:"completedCourses"`
	CurrentStreak     int     `json:"currentStreak"`
	AverageScore      float64 `json:"averageScore"`
	WeeklyGoalProgress float64 `json:"weeklyGoalProgress"`
	MonthlyGoalProgress float64 `json:"monthlyGoalProgress"`
}

// MockWeeklyActivity ?
type MockWeeklyActivity struct {
	Date       string `json:"date"`
	StudyTime  int    `json:"studyTime"`
	Exercises  int    `json:"exercises"`
	Score      int    `json:"score"`
}

// MockSkillProgress ?
type MockSkillProgress struct {
	Name     string  `json:"name"`
	Level    int     `json:"level"`
	Progress float64 `json:"progress"`
	Color    string  `json:"color"`
}

// MockRecommendation 
type MockRecommendation struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Type        string  `json:"type"`
	Priority    string  `json:"priority"`
	Confidence  float64 `json:"confidence"`
	Duration    int     `json:"estimatedDuration"`
	Icon        string  `json:"icon"`
}

// MockActivity 
type MockActivity struct {
	ID            string    `json:"id"`
	Date          time.Time `json:"date"`
	Type          string    `json:"type"`
	StudyTime     int       `json:"studyTime"`
	CoursesCount  int       `json:"coursesCount"`
	ExercisesCount int      `json:"exercisesCount"`
	Description   string    `json:"description"`
}

// MockAchievement 
type MockAchievement struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	Rarity      string    `json:"rarity"`
	UnlockedAt  time.Time `json:"unlockedAt"`
	Experience  int       `json:"experience"`
}

// GetMockLearnerProfile ?
func GetMockLearnerProfile() MockLearnerProfile {
	return MockLearnerProfile{
		ID:           uuid.New().String(),
		Name:         "",
		Level:        15,
		Experience:   12580,
		Avatar:       "https://api.dicebear.com/7.x/avataaars/svg?seed=zhang",
		JoinDate:     time.Now().AddDate(0, -6, 0),
		StudyStreak:  23,
		TotalStudyTime: 15600, // 
	}
}

// GetMockLearningAnalytics 
func GetMockLearningAnalytics() MockLearningAnalytics {
	return MockLearningAnalytics{
		TotalStudyTime:      15600,
		CompletedCourses:    42,
		CurrentStreak:       23,
		AverageScore:        87.5,
		WeeklyGoalProgress:  0.75,
		MonthlyGoalProgress: 0.68,
	}
}

// GetMockWeeklyActivity ?
func GetMockWeeklyActivity() []MockWeeklyActivity {
	activities := []MockWeeklyActivity{}
	now := time.Now()
	
	for i := 6; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		activities = append(activities, MockWeeklyActivity{
			Date:      date.Format("2006-01-02"),
			StudyTime: 60 + (i*15)%120,
			Exercises: 5 + i%8,
			Score:     80 + i%20,
		})
	}
	
	return activities
}

// GetMockSkillProgress ?
func GetMockSkillProgress() []MockSkillProgress {
	return []MockSkillProgress{
		{Name: "JavaScript", Level: 8, Progress: 0.85, Color: "#f7df1e"},
		{Name: "React", Level: 7, Progress: 0.72, Color: "#61dafb"},
		{Name: "Node.js", Level: 6, Progress: 0.65, Color: "#339933"},
		{Name: "TypeScript", Level: 5, Progress: 0.58, Color: "#3178c6"},
		{Name: "Python", Level: 4, Progress: 0.45, Color: "#3776ab"},
		{Name: "Go", Level: 3, Progress: 0.32, Color: "#00add8"},
	}
}

// GetMockRecommendations 
func GetMockRecommendations() []MockRecommendation {
	return []MockRecommendation{
		{
			ID:          uuid.New().String(),
			Title:       "React Hooks",
			Description: "JavaScriptReact Hooks?,
			Type:        "course",
			Priority:    "high",
			Confidence:  0.92,
			Duration:    180,
			Icon:        "book",
		},
		{
			ID:          uuid.New().String(),
			Title:       "TypeScript",
			Description: "ReactTypeScript?,
			Type:        "practice",
			Priority:    "medium",
			Confidence:  0.85,
			Duration:    120,
			Icon:        "code",
		},
		{
			ID:          uuid.New().String(),
			Title:       "Node.js",
			Description: "?,
			Type:        "tutorial",
			Priority:    "medium",
			Confidence:  0.78,
			Duration:    90,
			Icon:        "rocket",
		},
	}
}

// GetMockActivities 
func GetMockActivities() []MockActivity {
	activities := []MockActivity{}
	now := time.Now()
	
	for i := 0; i < 10; i++ {
		date := now.AddDate(0, 0, -i)
		activities = append(activities, MockActivity{
			ID:             uuid.New().String(),
			Date:           date,
			Type:           []string{"study", "practice", "review"}[i%3],
			StudyTime:      30 + (i*10)%90,
			CoursesCount:   1 + i%3,
			ExercisesCount: 3 + i%7,
			Description:    []string{"React", "JavaScript㷨", "TypeScript"}[i%3],
		})
	}
	
	return activities
}

// GetMockAchievements 
func GetMockAchievements() []MockAchievement {
	return []MockAchievement{
		{
			ID:          uuid.New().String(),
			Name:        "",
			Description: "21?,
			Icon:        "fire",
			Rarity:      "epic",
			UnlockedAt:  time.Now().AddDate(0, 0, -2),
			Experience:  500,
		},
		{
			ID:          uuid.New().String(),
			Name:        "JavaScript",
			Description: "JavaScript",
			Icon:        "crown",
			Rarity:      "legendary",
			UnlockedAt:  time.Now().AddDate(0, 0, -7),
			Experience:  1000,
		},
		{
			ID:          uuid.New().String(),
			Name:        "",
			Description: "??,
			Icon:        "sun",
			Rarity:      "rare",
			UnlockedAt:  time.Now().AddDate(0, 0, -1),
			Experience:  200,
		},
		{
			ID:          uuid.New().String(),
			Name:        "",
			Description: "100",
			Icon:        "target",
			Rarity:      "common",
			UnlockedAt:  time.Now().AddDate(0, 0, -14),
			Experience:  100,
		},
	}
}

