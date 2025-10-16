# 学习服务API文档

## 1. 服务概述

学习服务是太上老君AI平台的核心教育模块，基于S×C×T三轴理论设计，提供个性化学习路径、智能课程推荐、学习进度跟踪、知识评估、社交学习等功能，致力于打造智慧化的学习生态系统。

### 1.1 服务架构

```mermaid
graph TB
    subgraph "学习服务架构"
        subgraph "API接口层"
            REST[REST API]
            GraphQL[GraphQL API]
            WS[WebSocket API]
            WEBHOOK[Webhook API]
        end
        
        subgraph "学习引擎层"
            ADAPTIVE[自适应学习]
            RECOMMEND[智能推荐]
            ASSESS[智能评估]
            PROGRESS[进度跟踪]
            SOCIAL[社交学习]
        end
        
        subgraph "内容管理层"
            COURSE[课程管理]
            RESOURCE[资源管理]
            KNOWLEDGE[知识图谱]
            CURRICULUM[课程体系]
        end
        
        subgraph "分析引擎层"
            LEARNING_ANALYTICS[学习分析]
            BEHAVIOR_ANALYSIS[行为分析]
            PERFORMANCE[性能分析]
            PREDICTION[预测分析]
        end
        
        subgraph "数据层"
            USER_DATA[(用户数据)]
            LEARNING_DATA[(学习数据)]
            CONTENT_DATA[(内容数据)]
            ANALYTICS_DATA[(分析数据)]
        end
    end
    
    REST --> ADAPTIVE
    GraphQL --> RECOMMEND
    WS --> ASSESS
    WEBHOOK --> PROGRESS
    
    ADAPTIVE --> COURSE
    RECOMMEND --> RESOURCE
    ASSESS --> KNOWLEDGE
    PROGRESS --> CURRICULUM
    SOCIAL --> LEARNING_ANALYTICS
    
    COURSE --> USER_DATA
    RESOURCE --> LEARNING_DATA
    KNOWLEDGE --> CONTENT_DATA
    CURRICULUM --> ANALYTICS_DATA
```

### 1.2 核心功能

- **个性化学习**：基于用户特征的自适应学习路径
- **智能推荐**：课程、资源、学习伙伴推荐
- **学习跟踪**：实时学习进度和成果跟踪
- **智能评估**：多维度学习效果评估
- **社交学习**：学习社区、小组学习、师生互动
- **知识管理**：结构化知识体系和图谱
- **学习分析**：深度学习数据分析和洞察
- **移动学习**：跨平台学习体验

## 2. 课程管理

### 2.1 获取课程列表

```http
GET /api/v1/learning/courses?category=philosophy&level=beginner&limit=20&offset=0
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "courses": [
      {
        "course_id": "course_1234567890abcdef",
        "title": "道德经入门",
        "description": "深入浅出地学习老子的智慧",
        "category": "philosophy",
        "subcategory": "taoism",
        "level": "beginner",
        "duration": 1800,
        "lesson_count": 12,
        "instructor": {
          "id": "instructor_001",
          "name": "张教授",
          "avatar": "https://cdn.taishanglaojun.com/avatars/instructor_001.jpg",
          "credentials": ["哲学博士", "道教研究专家"]
        },
        "rating": {
          "average": 4.8,
          "count": 1250
        },
        "enrollment": {
          "current": 5680,
          "capacity": 10000
        },
        "pricing": {
          "type": "free",
          "original_price": 0,
          "current_price": 0
        },
        "tags": ["道德经", "老子", "哲学", "传统文化"],
        "thumbnail": "https://cdn.taishanglaojun.com/courses/course_1234567890abcdef_thumb.jpg",
        "preview_video": "https://cdn.taishanglaojun.com/videos/course_1234567890abcdef_preview.mp4",
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-15T10:30:00Z",
        "status": "published"
      }
    ],
    "pagination": {
      "total": 156,
      "page": 1,
      "per_page": 20,
      "total_pages": 8
    },
    "filters": {
      "categories": ["philosophy", "health", "culture", "meditation"],
      "levels": ["beginner", "intermediate", "advanced"],
      "durations": ["short", "medium", "long"],
      "pricing": ["free", "paid", "premium"]
    }
  }
}
```

### 2.2 获取课程详情

```http
GET /api/v1/learning/courses/{course_id}
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "course_id": "course_1234567890abcdef",
    "title": "道德经入门",
    "description": "深入浅出地学习老子的智慧，理解道家思想的核心理念",
    "long_description": "本课程将带您深入了解道德经的精髓...",
    "category": "philosophy",
    "level": "beginner",
    "duration": 1800,
    "language": "zh-CN",
    "instructor": {
      "id": "instructor_001",
      "name": "张教授",
      "bio": "哲学博士，专注道教研究20年",
      "avatar": "https://cdn.taishanglaojun.com/avatars/instructor_001.jpg"
    },
    "curriculum": {
      "chapters": [
        {
          "chapter_id": "chapter_001",
          "title": "道德经概述",
          "description": "了解道德经的历史背景和基本结构",
          "duration": 300,
          "lessons": [
            {
              "lesson_id": "lesson_001",
              "title": "老子其人其书",
              "type": "video",
              "duration": 180,
              "is_free": true,
              "resources": [
                {
                  "type": "video",
                  "url": "https://cdn.taishanglaojun.com/videos/lesson_001.mp4",
                  "quality": "1080p"
                },
                {
                  "type": "pdf",
                  "url": "https://cdn.taishanglaojun.com/materials/lesson_001.pdf",
                  "title": "课程讲义"
                }
              ]
            }
          ]
        }
      ]
    },
    "learning_objectives": [
      "理解道德经的基本思想",
      "掌握道家哲学的核心概念",
      "能够运用道家智慧指导生活"
    ],
    "prerequisites": [],
    "target_audience": ["哲学爱好者", "传统文化学习者", "人生智慧探索者"],
    "certification": {
      "available": true,
      "type": "completion_certificate",
      "requirements": {
        "completion_rate": 80,
        "quiz_score": 70
      }
    },
    "enrollment_info": {
      "is_enrolled": false,
      "enrollment_count": 5680,
      "capacity": 10000,
      "enrollment_deadline": "2024-12-31T23:59:59Z"
    },
    "reviews": {
      "average_rating": 4.8,
      "total_reviews": 1250,
      "rating_distribution": {
        "5": 850,
        "4": 300,
        "3": 80,
        "2": 15,
        "1": 5
      }
    }
  }
}
```

### 2.3 课程注册

```http
POST /api/v1/learning/courses/{course_id}/enroll
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "enrollment_type": "individual",
  "payment_method": "free",
  "learning_goals": [
    "understand_taoism_basics",
    "apply_wisdom_to_life",
    "complete_certification"
  ],
  "expected_completion_date": "2024-06-30",
  "study_schedule": {
    "hours_per_week": 3,
    "preferred_days": ["monday", "wednesday", "friday"],
    "preferred_time": "evening"
  }
}
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "enrollment_id": "enrollment_1234567890abcdef",
    "course_id": "course_1234567890abcdef",
    "user_id": "usr_1234567890abcdef",
    "enrollment_date": "2024-01-15T10:30:00Z",
    "status": "active",
    "progress": {
      "completion_percentage": 0,
      "current_lesson": null,
      "lessons_completed": 0,
      "total_lessons": 12
    },
    "learning_path": {
      "estimated_completion_date": "2024-06-30",
      "recommended_schedule": {
        "sessions_per_week": 3,
        "minutes_per_session": 60
      }
    },
    "access_info": {
      "access_granted": true,
      "expires_at": "2024-12-31T23:59:59Z",
      "download_allowed": true,
      "offline_access": true
    }
  }
}
```

### 2.4 课程进度更新

```http
PUT /api/v1/learning/courses/{course_id}/progress
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "lesson_id": "lesson_001",
  "action": "complete",
  "progress_data": {
    "watch_time": 180,
    "completion_percentage": 100,
    "quiz_score": 85,
    "notes": "很有启发的一课，对老子的生平有了更深的了解"
  },
  "learning_context": {
    "device_type": "mobile",
    "session_duration": 200,
    "interruptions": 1
  }
}
```

## 3. 学习路径管理

### 3.1 获取推荐学习路径

```http
POST /api/v1/learning/paths/recommend
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "learning_goals": [
    "master_chinese_philosophy",
    "improve_meditation_skills",
    "understand_traditional_medicine"
  ],
  "current_knowledge": {
    "philosophy": "beginner",
    "meditation": "intermediate",
    "medicine": "novice"
  },
  "time_constraints": {
    "available_hours_per_week": 10,
    "target_completion_months": 6,
    "preferred_schedule": "flexible"
  },
  "learning_preferences": {
    "content_types": ["video", "interactive", "reading"],
    "difficulty_progression": "gradual",
    "social_learning": true
  }
}
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "recommended_paths": [
      {
        "path_id": "path_1234567890abcdef",
        "title": "中国哲学与养生智慧",
        "description": "从哲学基础到实践应用的完整学习路径",
        "duration": {
          "estimated_weeks": 24,
          "hours_per_week": 8,
          "total_hours": 192
        },
        "difficulty_progression": "beginner_to_advanced",
        "match_score": 0.95,
        "phases": [
          {
            "phase_id": "phase_001",
            "title": "哲学基础",
            "duration_weeks": 8,
            "courses": [
              {
                "course_id": "course_1234567890abcdef",
                "title": "道德经入门",
                "order": 1,
                "is_prerequisite": true
              },
              {
                "course_id": "course_2234567890abcdef",
                "title": "儒家思想概论",
                "order": 2,
                "is_prerequisite": false
              }
            ]
          },
          {
            "phase_id": "phase_002",
            "title": "冥想实践",
            "duration_weeks": 6,
            "courses": [
              {
                "course_id": "course_3234567890abcdef",
                "title": "道家冥想技法",
                "order": 1,
                "is_prerequisite": false
              }
            ]
          }
        ],
        "learning_outcomes": [
          "深入理解中国哲学核心思想",
          "掌握传统冥想技法",
          "了解中医养生基础理论"
        ],
        "certification": {
          "available": true,
          "title": "中国哲学与养生专家认证"
        }
      }
    ],
    "personalization_factors": {
      "knowledge_gaps": ["traditional_medicine_basics"],
      "strength_areas": ["meditation_practice"],
      "recommended_focus": "philosophy_foundation"
    }
  }
}
```

### 3.2 创建自定义学习路径

```http
POST /api/v1/learning/paths
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "title": "我的道家修行之路",
  "description": "个人定制的道家哲学与实践学习路径",
  "visibility": "private",
  "courses": [
    {
      "course_id": "course_1234567890abcdef",
      "order": 1,
      "start_date": "2024-02-01",
      "target_completion_date": "2024-03-15"
    },
    {
      "course_id": "course_2234567890abcdef",
      "order": 2,
      "start_date": "2024-03-16",
      "target_completion_date": "2024-05-01"
    }
  ],
  "milestones": [
    {
      "title": "完成道德经基础学习",
      "description": "理解道德经核心思想",
      "target_date": "2024-03-15",
      "courses": ["course_1234567890abcdef"]
    }
  ],
  "study_schedule": {
    "hours_per_week": 5,
    "preferred_days": ["saturday", "sunday"],
    "reminder_settings": {
      "enabled": true,
      "frequency": "daily",
      "time": "19:00"
    }
  }
}
```

### 3.3 学习路径进度跟踪

```http
GET /api/v1/learning/paths/{path_id}/progress
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "path_id": "path_1234567890abcdef",
    "overall_progress": {
      "completion_percentage": 35.5,
      "courses_completed": 1,
      "total_courses": 3,
      "hours_studied": 68,
      "estimated_hours_remaining": 124
    },
    "current_phase": {
      "phase_id": "phase_002",
      "title": "冥想实践",
      "progress_percentage": 20,
      "estimated_completion": "2024-04-15"
    },
    "course_progress": [
      {
        "course_id": "course_1234567890abcdef",
        "title": "道德经入门",
        "status": "completed",
        "completion_date": "2024-03-10",
        "final_score": 88,
        "time_spent": 25
      },
      {
        "course_id": "course_2234567890abcdef",
        "title": "道家冥想技法",
        "status": "in_progress",
        "progress_percentage": 40,
        "current_lesson": "lesson_005",
        "time_spent": 15
      }
    ],
    "milestones": [
      {
        "milestone_id": "milestone_001",
        "title": "完成道德经基础学习",
        "status": "completed",
        "completion_date": "2024-03-10"
      }
    ],
    "achievements": [
      {
        "achievement_id": "achievement_001",
        "title": "哲学入门者",
        "description": "完成第一门哲学课程",
        "earned_date": "2024-03-10",
        "badge_url": "https://cdn.taishanglaojun.com/badges/philosophy_beginner.png"
      }
    ],
    "recommendations": [
      {
        "type": "study_schedule",
        "message": "建议增加每周学习时间到8小时以达到目标完成日期"
      },
      {
        "type": "supplementary_resource",
        "message": "推荐阅读《庄子》来加深对道家思想的理解",
        "resource_id": "resource_001"
      }
    ]
  }
}
```

## 4. 学习评估

### 4.1 创建测验

```http
POST /api/v1/learning/assessments/quizzes
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "title": "道德经第一章理解测验",
  "description": "测试对道德经开篇内容的理解",
  "course_id": "course_1234567890abcdef",
  "lesson_id": "lesson_001",
  "quiz_type": "formative",
  "time_limit": 1800,
  "attempts_allowed": 3,
  "passing_score": 70,
  "questions": [
    {
      "question_id": "q001",
      "type": "multiple_choice",
      "question": "道德经开篇"道可道，非常道"中的第一个"道"指的是什么？",
      "options": [
        {
          "id": "a",
          "text": "可以言说的道理"
        },
        {
          "id": "b", 
          "text": "永恒不变的道"
        },
        {
          "id": "c",
          "text": "道路"
        },
        {
          "id": "d",
          "text": "方法"
        }
      ],
      "correct_answer": "a",
      "explanation": "第一个'道'指的是可以用语言表达的道理或规律",
      "points": 10,
      "difficulty": "medium"
    },
    {
      "question_id": "q002",
      "type": "essay",
      "question": "请用自己的话解释"无为而治"的含义，并举一个现实生活中的例子。",
      "max_words": 200,
      "points": 20,
      "difficulty": "hard",
      "rubric": [
        {
          "criterion": "理解准确性",
          "points": 10,
          "description": "准确理解无为而治的含义"
        },
        {
          "criterion": "例子恰当性",
          "points": 10,
          "description": "举例恰当且与概念相符"
        }
      ]
    }
  ],
  "settings": {
    "randomize_questions": true,
    "show_correct_answers": true,
    "allow_review": true,
    "immediate_feedback": true
  }
}
```

### 4.2 提交测验答案

```http
POST /api/v1/learning/assessments/quizzes/{quiz_id}/submit
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "attempt_id": "attempt_1234567890abcdef",
  "answers": [
    {
      "question_id": "q001",
      "answer": "a",
      "time_spent": 45
    },
    {
      "question_id": "q002",
      "answer": "无为而治是指通过不过度干预来实现有效管理。例如，优秀的团队领导者往往通过营造良好的工作环境和文化，让团队成员自主发挥，而不是事无巨细地指挥每个细节。",
      "time_spent": 180
    }
  ],
  "submission_context": {
    "total_time_spent": 225,
    "device_type": "desktop",
    "browser": "chrome"
  }
}
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "submission_id": "submission_1234567890abcdef",
    "quiz_id": "quiz_1234567890abcdef",
    "attempt_id": "attempt_1234567890abcdef",
    "score": {
      "total_points": 25,
      "max_points": 30,
      "percentage": 83.3,
      "grade": "B+",
      "passed": true
    },
    "question_results": [
      {
        "question_id": "q001",
        "points_earned": 10,
        "points_possible": 10,
        "is_correct": true,
        "feedback": "正确！第一个'道'确实指可以言说的道理。"
      },
      {
        "question_id": "q002",
        "points_earned": 15,
        "points_possible": 20,
        "is_correct": true,
        "feedback": "很好的理解和例子，但可以更深入地解释无为的哲学内涵。",
        "detailed_feedback": {
          "理解准确性": {
            "points": 8,
            "max_points": 10,
            "comment": "基本理解正确，但缺少对'无为'深层含义的阐述"
          },
          "例子恰当性": {
            "points": 7,
            "max_points": 10,
            "comment": "例子恰当，但可以更具体地说明如何体现'无为'"
          }
        }
      }
    ],
    "completion_time": 225,
    "submitted_at": "2024-01-15T10:30:00Z",
    "next_steps": {
      "recommendations": [
        "复习无为而治的深层哲学含义",
        "阅读更多关于道家管理思想的资料"
      ],
      "next_lesson": "lesson_002"
    }
  }
}
```

### 4.3 技能评估

```http
POST /api/v1/learning/assessments/skills
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "assessment_type": "comprehensive",
  "skill_areas": [
    "chinese_philosophy_understanding",
    "meditation_practice",
    "traditional_culture_knowledge"
  ],
  "assessment_format": "adaptive",
  "duration_minutes": 60,
  "difficulty_range": {
    "min": "beginner",
    "max": "advanced"
  }
}
```

### 4.4 学习成果评估

```http
GET /api/v1/learning/assessments/outcomes/{user_id}?time_range=last_month
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "user_id": "usr_1234567890abcdef",
    "assessment_period": {
      "start_date": "2024-01-01T00:00:00Z",
      "end_date": "2024-01-31T23:59:59Z"
    },
    "overall_performance": {
      "learning_score": 85.5,
      "improvement_rate": 12.3,
      "consistency_score": 78.9,
      "engagement_level": "high"
    },
    "skill_assessments": [
      {
        "skill_area": "chinese_philosophy_understanding",
        "current_level": "intermediate",
        "proficiency_score": 82,
        "progress_from_start": 25,
        "strengths": [
          "道德经理解",
          "基本概念掌握"
        ],
        "improvement_areas": [
          "哲学思辨能力",
          "现代应用理解"
        ]
      }
    ],
    "learning_analytics": {
      "total_study_hours": 45.5,
      "courses_completed": 2,
      "quizzes_taken": 8,
      "average_quiz_score": 83.2,
      "learning_velocity": "above_average",
      "retention_rate": 89.5
    },
    "achievements": [
      {
        "achievement_id": "achievement_002",
        "title": "月度学习达人",
        "description": "本月学习时间超过40小时",
        "earned_date": "2024-01-31T23:59:59Z"
      }
    ],
    "recommendations": [
      {
        "type": "skill_development",
        "priority": "high",
        "message": "建议加强哲学思辨训练，可以参加讨论小组"
      },
      {
        "type": "learning_path",
        "priority": "medium",
        "message": "准备好进入高级课程学习"
      }
    ]
  }
}
```

## 5. 社交学习

### 5.1 创建学习小组

```http
POST /api/v1/learning/groups
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "name": "道德经研读小组",
  "description": "一起深入学习和讨论道德经的智慧",
  "type": "study_group",
  "privacy": "public",
  "max_members": 20,
  "course_id": "course_1234567890abcdef",
  "study_schedule": {
    "meeting_frequency": "weekly",
    "meeting_day": "sunday",
    "meeting_time": "19:00",
    "duration_minutes": 90
  },
  "group_rules": [
    "尊重每个成员的观点",
    "积极参与讨论",
    "按时完成学习任务"
  ],
  "tags": ["道德经", "哲学", "讨论", "传统文化"]
}
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "group_id": "group_1234567890abcdef",
    "name": "道德经研读小组",
    "description": "一起深入学习和讨论道德经的智慧",
    "type": "study_group",
    "privacy": "public",
    "creator": {
      "user_id": "usr_1234567890abcdef",
      "name": "张三",
      "role": "moderator"
    },
    "member_count": 1,
    "max_members": 20,
    "status": "active",
    "created_at": "2024-01-15T10:30:00Z",
    "invite_code": "DAODE2024",
    "group_url": "https://app.taishanglaojun.com/groups/group_1234567890abcdef"
  }
}
```

### 5.2 加入学习小组

```http
POST /api/v1/learning/groups/{group_id}/join
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "join_method": "invite_code",
  "invite_code": "DAODE2024",
  "introduction": "我对道德经很感兴趣，希望能和大家一起学习讨论",
  "learning_goals": [
    "深入理解道德经",
    "提高哲学思辨能力",
    "结交志同道合的朋友"
  ]
}
```

### 5.3 小组讨论

```http
POST /api/v1/learning/groups/{group_id}/discussions
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "title": "如何理解"上善若水"？",
  "content": "道德经第八章提到"上善若水"，大家是如何理解这句话的深层含义的？在现代生活中如何实践这种智慧？",
  "category": "philosophical_discussion",
  "tags": ["上善若水", "道德经第八章", "实践智慧"],
  "related_lesson": "lesson_008",
  "discussion_type": "open_ended",
  "attachments": [
    {
      "type": "image",
      "url": "https://example.com/water_image.jpg",
      "description": "水的形态变化图"
    }
  ]
}
```

### 5.4 获取小组活动

```http
GET /api/v1/learning/groups/{group_id}/activities?limit=20&offset=0
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "activities": [
      {
        "activity_id": "activity_1234567890abcdef",
        "type": "discussion",
        "title": "如何理解"上善若水"？",
        "author": {
          "user_id": "usr_1234567890abcdef",
          "name": "张三",
          "avatar": "https://cdn.taishanglaojun.com/avatars/usr_001.jpg"
        },
        "created_at": "2024-01-15T10:30:00Z",
        "stats": {
          "replies": 15,
          "likes": 8,
          "views": 45
        },
        "latest_reply": {
          "author": "李四",
          "content": "我觉得水的特质正是道家思想的体现...",
          "created_at": "2024-01-15T14:20:00Z"
        }
      },
      {
        "activity_id": "activity_2234567890abcdef",
        "type": "study_session",
        "title": "第三章集体学习",
        "scheduled_at": "2024-01-21T19:00:00Z",
        "duration_minutes": 90,
        "participants_count": 12,
        "max_participants": 20,
        "status": "scheduled"
      }
    ]
  }
}
```

### 5.5 学习伙伴匹配

```http
POST /api/v1/learning/partners/match
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "matching_criteria": {
    "learning_goals": ["chinese_philosophy", "meditation"],
    "skill_level": "beginner_to_intermediate",
    "study_schedule": {
      "preferred_days": ["weekends"],
      "preferred_time": "evening",
      "hours_per_week": 5
    },
    "personality_traits": ["patient", "curious", "collaborative"],
    "communication_style": "discussion_oriented"
  },
  "preferences": {
    "age_range": {
      "min": 25,
      "max": 45
    },
    "location_preference": "any",
    "language": "zh-CN",
    "max_matches": 5
  }
}
```

## 6. 学习资源管理

### 6.1 获取学习资源

```http
GET /api/v1/learning/resources?type=video&category=philosophy&level=beginner
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "resources": [
      {
        "resource_id": "resource_1234567890abcdef",
        "title": "道德经朗读版",
        "description": "专业朗读的道德经全文，有助于理解和记忆",
        "type": "audio",
        "category": "philosophy",
        "subcategory": "taoism",
        "level": "beginner",
        "duration": 3600,
        "file_size": 52428800,
        "format": "mp3",
        "quality": "high",
        "url": "https://cdn.taishanglaojun.com/audio/daodejing_reading.mp3",
        "download_url": "https://cdn.taishanglaojun.com/downloads/daodejing_reading.mp3",
        "thumbnail": "https://cdn.taishanglaojun.com/thumbnails/audio_daodejing.jpg",
        "author": {
          "name": "王教授",
          "credentials": ["古典文学博士", "朗诵艺术家"]
        },
        "metadata": {
          "chapters": 81,
          "language": "zh-CN",
          "accent": "standard_mandarin",
          "background_music": false
        },
        "tags": ["道德经", "朗读", "古典", "哲学"],
        "rating": {
          "average": 4.7,
          "count": 892
        },
        "usage_rights": {
          "download_allowed": true,
          "offline_access": true,
          "sharing_allowed": false,
          "commercial_use": false
        },
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-10T15:30:00Z"
      }
    ],
    "filters": {
      "types": ["video", "audio", "pdf", "interactive", "quiz"],
      "categories": ["philosophy", "health", "culture", "meditation"],
      "levels": ["beginner", "intermediate", "advanced"],
      "formats": ["mp4", "mp3", "pdf", "html5"]
    }
  }
}
```

### 6.2 上传学习资源

```http
POST /api/v1/learning/resources
Content-Type: multipart/form-data
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

title: "我的道德经学习笔记"
description: "个人整理的道德经重点章节笔记和心得"
type: "document"
category: "philosophy"
level: "intermediate"
file: [binary file data]
tags: ["道德经", "笔记", "心得", "个人整理"]
visibility: "public"
license: "cc_by_sa"
```

### 6.3 收藏学习资源

```http
POST /api/v1/learning/resources/{resource_id}/bookmark
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "collection_name": "道德经学习资料",
  "notes": "很好的朗读版本，适合冥想时聆听",
  "tags": ["收藏", "朗读", "冥想"]
}
```

### 6.4 资源使用统计

```http
POST /api/v1/learning/resources/{resource_id}/usage
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "action": "view",
  "duration": 1800,
  "completion_percentage": 75,
  "device_type": "mobile",
  "context": {
    "course_id": "course_1234567890abcdef",
    "lesson_id": "lesson_003",
    "study_session_id": "session_1234567890abcdef"
  }
}
```

## 7. 学习分析与洞察

### 7.1 个人学习分析

```http
GET /api/v1/learning/analytics/personal?time_range=last_3_months
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "user_id": "usr_1234567890abcdef",
    "analysis_period": {
      "start_date": "2023-11-01T00:00:00Z",
      "end_date": "2024-01-31T23:59:59Z"
    },
    "learning_summary": {
      "total_study_hours": 156.5,
      "courses_enrolled": 5,
      "courses_completed": 3,
      "lessons_completed": 45,
      "quizzes_taken": 23,
      "average_quiz_score": 84.2,
      "certificates_earned": 2
    },
    "learning_patterns": {
      "preferred_study_times": [
        {
          "time_slot": "19:00-21:00",
          "frequency": 0.65,
          "average_duration": 90
        },
        {
          "time_slot": "09:00-11:00",
          "frequency": 0.35,
          "average_duration": 120
        }
      ],
      "study_frequency": {
        "daily_average": 1.2,
        "weekly_pattern": {
          "monday": 0.8,
          "tuesday": 0.9,
          "wednesday": 1.1,
          "thursday": 0.7,
          "friday": 0.6,
          "saturday": 1.8,
          "sunday": 2.1
        }
      },
      "device_usage": {
        "mobile": 0.45,
        "desktop": 0.35,
        "tablet": 0.20
      }
    },
    "performance_trends": {
      "learning_velocity": {
        "current": "above_average",
        "trend": "improving",
        "change_percentage": 15.3
      },
      "knowledge_retention": {
        "short_term": 0.89,
        "long_term": 0.76,
        "trend": "stable"
      },
      "engagement_level": {
        "current": "high",
        "trend": "increasing",
        "factors": ["interactive_content", "social_learning"]
      }
    },
    "skill_development": [
      {
        "skill": "chinese_philosophy",
        "current_level": "intermediate",
        "progress_rate": 0.23,
        "proficiency_score": 78,
        "improvement_areas": ["critical_thinking", "modern_application"]
      },
      {
        "skill": "meditation_practice",
        "current_level": "advanced_beginner",
        "progress_rate": 0.31,
        "proficiency_score": 65,
        "improvement_areas": ["consistency", "advanced_techniques"]
      }
    ],
    "learning_preferences": {
      "content_types": {
        "video": 0.45,
        "interactive": 0.25,
        "reading": 0.20,
        "audio": 0.10
      },
      "difficulty_preference": "gradual_progression",
      "social_learning_engagement": "moderate"
    },
    "recommendations": [
      {
        "type": "study_schedule",
        "priority": "high",
        "message": "建议保持周末的高强度学习，并在工作日增加短时间复习"
      },
      {
        "type": "content_recommendation",
        "priority": "medium",
        "message": "基于您的进度，推荐开始学习高级哲学思辨课程"
      }
    ]
  }
}
```

### 7.2 学习效果预测

```http
POST /api/v1/learning/analytics/predictions
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "prediction_type": "course_completion",
  "course_id": "course_1234567890abcdef",
  "prediction_horizon": "30_days",
  "factors": [
    "current_progress",
    "study_frequency",
    "quiz_performance",
    "engagement_level"
  ]
}
```

### 7.3 学习群体分析

```http
GET /api/v1/learning/analytics/cohort?course_id=course_1234567890abcdef&start_date=2024-01-01
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

## 8. 学习提醒与通知

### 8.1 设置学习提醒

```http
POST /api/v1/learning/reminders
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "reminder_type": "study_schedule",
  "title": "道德经学习时间",
  "message": "该开始今天的道德经学习了！",
  "schedule": {
    "frequency": "daily",
    "time": "19:00",
    "days": ["monday", "wednesday", "friday", "sunday"],
    "timezone": "Asia/Shanghai"
  },
  "conditions": {
    "only_if_behind_schedule": true,
    "skip_if_already_studied": true,
    "minimum_gap_hours": 4
  },
  "delivery_methods": ["push_notification", "email"],
  "course_id": "course_1234567890abcdef"
}
```

### 8.2 获取学习通知

```http
GET /api/v1/learning/notifications?status=unread&limit=20
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "notifications": [
      {
        "notification_id": "notif_1234567890abcdef",
        "type": "achievement",
        "title": "恭喜获得新成就！",
        "message": "您已完成"道德经入门"课程，获得"哲学启蒙者"徽章！",
        "priority": "medium",
        "status": "unread",
        "created_at": "2024-01-15T10:30:00Z",
        "data": {
          "achievement_id": "achievement_003",
          "badge_url": "https://cdn.taishanglaojun.com/badges/philosophy_enlightener.png"
        },
        "actions": [
          {
            "type": "view_achievement",
            "label": "查看成就",
            "url": "/achievements/achievement_003"
          }
        ]
      },
      {
        "notification_id": "notif_2234567890abcdef",
        "type": "study_reminder",
        "title": "学习提醒",
        "message": "您已经2天没有学习了，继续保持学习习惯吧！",
        "priority": "low",
        "status": "unread",
        "created_at": "2024-01-15T09:00:00Z",
        "data": {
          "course_id": "course_1234567890abcdef",
          "last_study_date": "2024-01-13T20:30:00Z"
        }
      }
    ],
    "unread_count": 5,
    "total_count": 23
  }
}
```

## 9. 移动学习支持

### 9.1 离线内容同步

```http
POST /api/v1/learning/offline/sync
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "sync_type": "selective",
  "content_items": [
    {
      "type": "course",
      "id": "course_1234567890abcdef",
      "quality": "medium",
      "include_resources": true
    },
    {
      "type": "lesson",
      "id": "lesson_001",
      "quality": "high",
      "include_subtitles": true
    }
  ],
  "device_info": {
    "storage_available": 2048,
    "connection_type": "wifi",
    "battery_level": 85
  }
}
```

### 9.2 学习进度同步

```http
POST /api/v1/learning/sync/progress
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "device_id": "device_1234567890abcdef",
  "sync_data": [
    {
      "course_id": "course_1234567890abcdef",
      "lesson_id": "lesson_001",
      "progress_percentage": 75,
      "time_spent": 1800,
      "last_position": 1350,
      "completed_at": null,
      "offline_timestamp": "2024-01-15T10:30:00Z"
    }
  ],
  "conflict_resolution": "server_wins"
}
```

## 10. 数据模型

### 10.1 课程数据模型

```typescript
interface Course {
  course_id: string;
  title: string;
  description: string;
  long_description?: string;
  category: string;
  subcategory?: string;
  level: 'beginner' | 'intermediate' | 'advanced';
  duration: number; // 秒
  language: string;
  instructor: Instructor;
  curriculum: Curriculum;
  learning_objectives: string[];
  prerequisites: string[];
  target_audience: string[];
  certification?: Certification;
  pricing: Pricing;
  rating: Rating;
  enrollment: EnrollmentInfo;
  tags: string[];
  thumbnail: string;
  preview_video?: string;
  status: 'draft' | 'published' | 'archived';
  created_at: string;
  updated_at: string;
}

interface Instructor {
  id: string;
  name: string;
  bio?: string;
  avatar?: string;
  credentials: string[];
  rating?: number;
}

interface Curriculum {
  chapters: Chapter[];
}

interface Chapter {
  chapter_id: string;
  title: string;
  description?: string;
  duration: number;
  lessons: Lesson[];
}

interface Lesson {
  lesson_id: string;
  title: string;
  type: 'video' | 'audio' | 'text' | 'interactive' | 'quiz';
  duration: number;
  is_free: boolean;
  resources: Resource[];
  learning_objectives?: string[];
}
```

### 10.2 学习进度数据模型

```typescript
interface LearningProgress {
  user_id: string;
  course_id: string;
  enrollment_date: string;
  status: 'active' | 'completed' | 'paused' | 'dropped';
  completion_percentage: number;
  current_lesson?: string;
  lessons_completed: number;
  total_lessons: number;
  time_spent: number; // 秒
  last_accessed: string;
  quiz_scores: QuizScore[];
  achievements: Achievement[];
  notes: Note[];
}

interface QuizScore {
  quiz_id: string;
  score: number;
  max_score: number;
  percentage: number;
  attempt_count: number;
  completed_at: string;
}

interface Achievement {
  achievement_id: string;
  title: string;
  description: string;
  badge_url: string;
  earned_date: string;
  category: string;
}
```

## 11. 错误处理

### 11.1 学习服务特定错误

```typescript
enum LearningServiceErrorCode {
  // 课程错误
  COURSE_NOT_FOUND = 'LEARN1001',
  COURSE_NOT_AVAILABLE = 'LEARN1002',
  COURSE_ACCESS_DENIED = 'LEARN1003',
  COURSE_ENROLLMENT_FULL = 'LEARN1004',
  
  // 学习进度错误
  INVALID_PROGRESS_DATA = 'LEARN2001',
  PROGRESS_SYNC_FAILED = 'LEARN2002',
  LESSON_NOT_ACCESSIBLE = 'LEARN2003',
  
  // 评估错误
  QUIZ_NOT_FOUND = 'LEARN3001',
  QUIZ_ALREADY_COMPLETED = 'LEARN3002',
  QUIZ_TIME_EXPIRED = 'LEARN3003',
  INVALID_ANSWER_FORMAT = 'LEARN3004',
  
  // 社交学习错误
  GROUP_NOT_FOUND = 'LEARN4001',
  GROUP_ACCESS_DENIED = 'LEARN4002',
  GROUP_FULL = 'LEARN4003',
  DISCUSSION_NOT_FOUND = 'LEARN4004',
  
  // 资源错误
  RESOURCE_NOT_FOUND = 'LEARN5001',
  RESOURCE_ACCESS_DENIED = 'LEARN5002',
  RESOURCE_UPLOAD_FAILED = 'LEARN5003',
  INVALID_RESOURCE_FORMAT = 'LEARN5004',
}
```

## 12. 性能优化

### 12.1 缓存策略

```yaml
# 学习服务缓存配置
cache_strategy:
  course_catalog:
    ttl: 3600
    key_pattern: "learning:courses:{category}:{level}"
    
  user_progress:
    ttl: 300
    key_pattern: "learning:progress:{user_id}:{course_id}"
    
  quiz_results:
    ttl: 86400
    key_pattern: "learning:quiz:{quiz_id}:{user_id}"
    
  learning_analytics:
    ttl: 1800
    key_pattern: "learning:analytics:{user_id}:{period}"
```

### 12.2 数据库优化

```sql
-- 学习进度查询优化索引
CREATE INDEX idx_learning_progress_user_course 
ON learning_progress(user_id, course_id);

CREATE INDEX idx_learning_progress_status_updated 
ON learning_progress(status, updated_at);

-- 课程搜索优化索引
CREATE INDEX idx_courses_category_level_status 
ON courses(category, level, status);

CREATE INDEX idx_courses_fulltext 
ON courses USING gin(to_tsvector('chinese', title || ' ' || description));
```

## 13. 监控与指标

### 13.1 关键指标

```yaml
# 学习服务监控指标
metrics:
  engagement:
    - name: "course_completion_rate"
      description: "课程完成率"
      target: "> 70%"
      
    - name: "daily_active_learners"
      description: "日活跃学习者"
      
    - name: "average_study_time"
      description: "平均学习时长"
      target: "> 30min/day"
      
  quality:
    - name: "quiz_pass_rate"
      description: "测验通过率"
      target: "> 80%"
      
    - name: "content_rating"
      description: "内容评分"
      target: "> 4.0/5.0"
      
  performance:
    - name: "api_response_time"
      description: "API响应时间"
      target: "< 500ms"
      
    - name: "video_loading_time"
      description: "视频加载时间"
      target: "< 3s"
```

## 14. 相关文档

- [API概览文档](../api-overview.md)
- [用户服务API](./user-service-api.md)
- [AI服务API](./ai-service-api.md)
- [学习分析指南](../learning-analytics-guide.md)
- [移动学习最佳实践](../mobile-learning-guide.md)