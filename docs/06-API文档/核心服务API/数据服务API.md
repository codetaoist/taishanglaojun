# 数据服务API文档

## 1. 服务概述

数据服务是太上老君AI平台的核心数据管理模块，基于S×C×T三轴理论设计，提供数据采集、存储、处理、分析、可视化等全生命周期数据管理服务，支持多种数据源和数据格式，确保数据的安全性、一致性和可用性。

### 1.1 服务架构

```mermaid
graph TB
    subgraph "数据服务架构"
        subgraph "API网关层"
            REST[REST API]
            GraphQL[GraphQL API]
            STREAM[流式API]
            BATCH[批处理API]
        end
        
        subgraph "数据处理层"
            INGESTION[数据摄取]
            VALIDATION[数据验证]
            TRANSFORMATION[数据转换]
            ENRICHMENT[数据增强]
        end
        
        subgraph "数据存储层"
            RELATIONAL[(关系数据库)]
            DOCUMENT[(文档数据库)]
            GRAPH[(图数据库)]
            VECTOR[(向量数据库)]
            CACHE[(缓存层)]
        end
        
        subgraph "数据分析层"
            ANALYTICS[数据分析]
            ML_PIPELINE[机器学习管道]
            REPORTING[报表生成]
            VISUALIZATION[数据可视化]
        end
        
        subgraph "数据治理层"
            CATALOG[数据目录]
            LINEAGE[数据血缘]
            QUALITY[数据质量]
            SECURITY[数据安全]
        end
    end
    
    REST --> INGESTION
    GraphQL --> VALIDATION
    STREAM --> TRANSFORMATION
    BATCH --> ENRICHMENT
    
    INGESTION --> RELATIONAL
    VALIDATION --> DOCUMENT
    TRANSFORMATION --> GRAPH
    ENRICHMENT --> VECTOR
    
    RELATIONAL --> ANALYTICS
    DOCUMENT --> ML_PIPELINE
    GRAPH --> REPORTING
    VECTOR --> VISUALIZATION
    
    ANALYTICS --> CATALOG
    ML_PIPELINE --> LINEAGE
    REPORTING --> QUALITY
    VISUALIZATION --> SECURITY
```

### 1.2 核心功能

- **数据摄取**：支持批量和实时数据摄取
- **数据存储**：多模态数据存储和管理
- **数据处理**：ETL/ELT数据处理管道
- **数据分析**：统计分析和机器学习
- **数据可视化**：图表和仪表板生成
- **数据治理**：数据质量、安全和合规
- **数据API**：统一数据访问接口
- **实时流处理**：流式数据处理和分析

## 2. 数据摄取服务

### 2.1 批量数据上传

```http
POST /api/v1/data/ingestion/batch
Content-Type: multipart/form-data
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

file: [binary file data]
dataset_name: "user_learning_records"
format: "csv"
schema_id: "schema_1234567890abcdef"
options: {
  "delimiter": ",",
  "header": true,
  "encoding": "utf-8",
  "skip_rows": 0
}
metadata: {
  "source": "learning_platform",
  "collection_date": "2024-01-15",
  "description": "用户学习记录数据"
}
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "ingestion_id": "ingestion_1234567890abcdef",
    "dataset_name": "user_learning_records",
    "status": "processing",
    "file_info": {
      "filename": "learning_records_20240115.csv",
      "size": 10485760,
      "format": "csv",
      "rows_estimated": 50000
    },
    "validation_results": {
      "schema_valid": true,
      "data_quality_score": 0.95,
      "issues": [
        {
          "type": "warning",
          "message": "发现3行数据缺少可选字段'notes'",
          "count": 3
        }
      ]
    },
    "processing_info": {
      "estimated_completion": "2024-01-15T10:35:00Z",
      "progress_url": "/api/v1/data/ingestion/ingestion_1234567890abcdef/progress"
    },
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

### 2.2 实时数据流摄取

```http
POST /api/v1/data/ingestion/stream
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "stream_name": "user_activity_stream",
  "data_source": {
    "type": "kafka",
    "config": {
      "bootstrap_servers": ["kafka1:9092", "kafka2:9092"],
      "topic": "user_activities",
      "consumer_group": "data_service_group"
    }
  },
  "schema_id": "schema_2234567890abcdef",
  "processing_config": {
    "batch_size": 1000,
    "batch_timeout": 5000,
    "parallelism": 4,
    "checkpoint_interval": 60000
  },
  "output_config": {
    "primary_storage": "postgresql",
    "secondary_storage": "elasticsearch",
    "cache_storage": "redis"
  }
}
```

### 2.3 API数据摄取

```http
POST /api/v1/data/ingestion/api
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "source_api": {
    "url": "https://external-api.example.com/data",
    "method": "GET",
    "headers": {
      "Authorization": "Bearer external_token",
      "Content-Type": "application/json"
    },
    "pagination": {
      "type": "offset",
      "page_size": 100,
      "max_pages": 1000
    }
  },
  "schedule": {
    "frequency": "hourly",
    "start_time": "2024-01-15T11:00:00Z",
    "timezone": "Asia/Shanghai"
  },
  "data_mapping": {
    "user_id": "$.id",
    "username": "$.name",
    "email": "$.email",
    "created_at": "$.registration_date"
  },
  "dataset_name": "external_users"
}
```

### 2.4 获取摄取进度

```http
GET /api/v1/data/ingestion/{ingestion_id}/progress
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "ingestion_id": "ingestion_1234567890abcdef",
    "status": "completed",
    "progress": {
      "percentage": 100,
      "rows_processed": 49997,
      "rows_total": 50000,
      "rows_failed": 3,
      "processing_rate": 1250.5
    },
    "timing": {
      "started_at": "2024-01-15T10:30:00Z",
      "completed_at": "2024-01-15T10:34:30Z",
      "duration_seconds": 270
    },
    "results": {
      "rows_inserted": 49997,
      "rows_updated": 0,
      "rows_skipped": 3,
      "data_quality_score": 0.9999
    },
    "errors": [
      {
        "row_number": 1523,
        "error_type": "validation_error",
        "message": "Invalid email format",
        "data": "user1523@invalid"
      }
    ]
  }
}
```

## 3. 数据查询服务

### 3.1 结构化数据查询

```http
POST /api/v1/data/query/sql
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "query": "SELECT user_id, course_id, completion_rate, study_time FROM learning_progress WHERE completion_rate > ? AND created_at >= ?",
  "parameters": [0.8, "2024-01-01"],
  "database": "learning_db",
  "options": {
    "limit": 1000,
    "timeout": 30,
    "format": "json",
    "include_metadata": true
  }
}
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "query_id": "query_1234567890abcdef",
    "results": [
      {
        "user_id": "usr_1234567890abcdef",
        "course_id": "course_1234567890abcdef",
        "completion_rate": 0.95,
        "study_time": 7200
      },
      {
        "user_id": "usr_2234567890abcdef",
        "course_id": "course_1234567890abcdef",
        "completion_rate": 0.87,
        "study_time": 5400
      }
    ],
    "metadata": {
      "total_rows": 2,
      "execution_time": 0.15,
      "columns": [
        {
          "name": "user_id",
          "type": "varchar",
          "nullable": false
        },
        {
          "name": "course_id",
          "type": "varchar",
          "nullable": false
        },
        {
          "name": "completion_rate",
          "type": "decimal",
          "nullable": true
        },
        {
          "name": "study_time",
          "type": "integer",
          "nullable": true
        }
      ]
    },
    "pagination": {
      "page": 1,
      "per_page": 1000,
      "total_pages": 1,
      "has_more": false
    }
  }
}
```

### 3.2 文档数据查询

```http
POST /api/v1/data/query/document
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "collection": "user_profiles",
  "filter": {
    "age": {"$gte": 25, "$lte": 45},
    "interests": {"$in": ["philosophy", "meditation"]},
    "location.country": "China"
  },
  "projection": {
    "user_id": 1,
    "name": 1,
    "age": 1,
    "interests": 1,
    "learning_preferences": 1
  },
  "sort": {"created_at": -1},
  "limit": 50,
  "skip": 0
}
```

### 3.3 图数据查询

```http
POST /api/v1/data/query/graph
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "query_type": "cypher",
  "query": "MATCH (u:User)-[r:ENROLLED_IN]->(c:Course) WHERE c.category = $category RETURN u.name, c.title, r.enrollment_date",
  "parameters": {
    "category": "philosophy"
  },
  "options": {
    "limit": 100,
    "include_stats": true
  }
}
```

### 3.4 向量相似度查询

```http
POST /api/v1/data/query/vector
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "collection": "course_embeddings",
  "vector": [0.1, 0.2, 0.3, ...],
  "similarity_metric": "cosine",
  "top_k": 10,
  "filter": {
    "category": "philosophy",
    "level": ["beginner", "intermediate"]
  },
  "include_metadata": true,
  "include_vectors": false
}
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "query_id": "vector_query_1234567890abcdef",
    "results": [
      {
        "id": "course_1234567890abcdef",
        "score": 0.95,
        "metadata": {
          "title": "道德经入门",
          "category": "philosophy",
          "level": "beginner",
          "description": "深入浅出地学习老子的智慧"
        }
      },
      {
        "id": "course_2234567890abcdef",
        "score": 0.87,
        "metadata": {
          "title": "儒家思想概论",
          "category": "philosophy",
          "level": "intermediate",
          "description": "系统学习儒家核心思想"
        }
      }
    ],
    "query_metadata": {
      "total_results": 2,
      "search_time": 0.05,
      "collection_size": 10000
    }
  }
}
```

## 4. 数据处理服务

### 4.1 创建数据处理管道

```http
POST /api/v1/data/pipelines
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "name": "user_behavior_analysis_pipeline",
  "description": "用户行为数据分析处理管道",
  "source": {
    "type": "database",
    "config": {
      "database": "analytics_db",
      "table": "user_activities",
      "query": "SELECT * FROM user_activities WHERE created_at >= NOW() - INTERVAL '1 DAY'"
    }
  },
  "transformations": [
    {
      "type": "filter",
      "config": {
        "condition": "activity_type IN ('course_view', 'lesson_complete', 'quiz_submit')"
      }
    },
    {
      "type": "aggregate",
      "config": {
        "group_by": ["user_id", "activity_type"],
        "aggregations": {
          "activity_count": "COUNT(*)",
          "total_duration": "SUM(duration)",
          "avg_duration": "AVG(duration)"
        }
      }
    },
    {
      "type": "enrich",
      "config": {
        "lookup_table": "users",
        "join_key": "user_id",
        "fields": ["name", "age", "location"]
      }
    }
  ],
  "destination": {
    "type": "database",
    "config": {
      "database": "analytics_db",
      "table": "user_behavior_summary",
      "write_mode": "upsert"
    }
  },
  "schedule": {
    "frequency": "daily",
    "time": "02:00",
    "timezone": "Asia/Shanghai"
  },
  "monitoring": {
    "alerts": {
      "failure": true,
      "data_quality_threshold": 0.95,
      "processing_time_threshold": 3600
    }
  }
}
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "pipeline_id": "pipeline_1234567890abcdef",
    "name": "user_behavior_analysis_pipeline",
    "status": "created",
    "version": 1,
    "created_at": "2024-01-15T10:30:00Z",
    "next_run": "2024-01-16T02:00:00Z",
    "estimated_runtime": 1800,
    "resource_requirements": {
      "cpu": "2 cores",
      "memory": "4GB",
      "storage": "10GB"
    }
  }
}
```

### 4.2 执行数据处理管道

```http
POST /api/v1/data/pipelines/{pipeline_id}/execute
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "execution_type": "manual",
  "parameters": {
    "start_date": "2024-01-15",
    "end_date": "2024-01-15",
    "batch_size": 10000
  },
  "options": {
    "dry_run": false,
    "parallel_execution": true,
    "checkpoint_enabled": true
  }
}
```

### 4.3 获取管道执行状态

```http
GET /api/v1/data/pipelines/{pipeline_id}/executions/{execution_id}
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "execution_id": "execution_1234567890abcdef",
    "pipeline_id": "pipeline_1234567890abcdef",
    "status": "running",
    "progress": {
      "percentage": 65,
      "current_stage": "aggregate",
      "stages_completed": 2,
      "total_stages": 4
    },
    "timing": {
      "started_at": "2024-01-15T10:30:00Z",
      "estimated_completion": "2024-01-15T11:00:00Z",
      "elapsed_seconds": 1200
    },
    "metrics": {
      "records_processed": 650000,
      "records_total": 1000000,
      "processing_rate": 541.67,
      "memory_usage": "2.1GB",
      "cpu_usage": "75%"
    },
    "stage_details": [
      {
        "stage": "filter",
        "status": "completed",
        "records_in": 1000000,
        "records_out": 850000,
        "duration_seconds": 300
      },
      {
        "stage": "aggregate",
        "status": "running",
        "records_in": 850000,
        "records_out": 0,
        "duration_seconds": 900
      }
    ]
  }
}
```

### 4.4 数据清洗服务

```http
POST /api/v1/data/cleaning
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "dataset": "user_learning_records",
  "cleaning_rules": [
    {
      "type": "remove_duplicates",
      "config": {
        "columns": ["user_id", "course_id", "lesson_id"],
        "keep": "last"
      }
    },
    {
      "type": "handle_missing_values",
      "config": {
        "strategy": "fill",
        "columns": {
          "study_time": {"method": "median"},
          "completion_rate": {"method": "forward_fill"},
          "notes": {"method": "constant", "value": ""}
        }
      }
    },
    {
      "type": "outlier_detection",
      "config": {
        "method": "iqr",
        "columns": ["study_time", "completion_rate"],
        "action": "flag"
      }
    },
    {
      "type": "data_validation",
      "config": {
        "rules": [
          {
            "column": "completion_rate",
            "condition": "BETWEEN 0 AND 1"
          },
          {
            "column": "study_time",
            "condition": "> 0"
          }
        ]
      }
    }
  ],
  "output_dataset": "user_learning_records_cleaned"
}
```

## 5. 数据分析服务

### 5.1 统计分析

```http
POST /api/v1/data/analytics/statistics
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "dataset": "user_learning_records",
  "analysis_type": "descriptive",
  "columns": ["study_time", "completion_rate", "quiz_score"],
  "group_by": ["course_category", "user_age_group"],
  "statistics": [
    "count",
    "mean",
    "median",
    "std",
    "min",
    "max",
    "percentiles"
  ],
  "filters": {
    "created_at": {
      "gte": "2024-01-01",
      "lte": "2024-01-31"
    }
  }
}
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "analysis_id": "analysis_1234567890abcdef",
    "dataset": "user_learning_records",
    "results": {
      "overall": {
        "study_time": {
          "count": 10000,
          "mean": 3600.5,
          "median": 3200.0,
          "std": 1200.3,
          "min": 300,
          "max": 14400,
          "percentiles": {
            "25": 2400,
            "50": 3200,
            "75": 4800,
            "90": 6000,
            "95": 7200
          }
        },
        "completion_rate": {
          "count": 10000,
          "mean": 0.75,
          "median": 0.80,
          "std": 0.25,
          "min": 0.0,
          "max": 1.0
        }
      },
      "grouped": [
        {
          "group": {
            "course_category": "philosophy",
            "user_age_group": "25-35"
          },
          "statistics": {
            "study_time": {
              "count": 2500,
              "mean": 4200.8,
              "median": 3800.0
            }
          }
        }
      ]
    },
    "metadata": {
      "execution_time": 2.5,
      "data_points": 10000,
      "generated_at": "2024-01-15T10:30:00Z"
    }
  }
}
```

### 5.2 趋势分析

```http
POST /api/v1/data/analytics/trends
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "dataset": "user_activities",
  "time_column": "created_at",
  "value_column": "activity_count",
  "time_range": {
    "start": "2024-01-01",
    "end": "2024-01-31"
  },
  "granularity": "daily",
  "group_by": ["activity_type"],
  "analysis_options": {
    "detect_seasonality": true,
    "detect_anomalies": true,
    "forecast_periods": 7,
    "confidence_interval": 0.95
  }
}
```

### 5.3 相关性分析

```http
POST /api/v1/data/analytics/correlation
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "dataset": "learning_outcomes",
  "variables": [
    "study_time",
    "completion_rate",
    "quiz_score",
    "engagement_level",
    "social_interaction"
  ],
  "method": "pearson",
  "significance_level": 0.05,
  "visualization": {
    "type": "heatmap",
    "include_p_values": true
  }
}
```

### 5.4 机器学习分析

```http
POST /api/v1/data/analytics/ml
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "task_type": "classification",
  "dataset": "user_behavior_features",
  "target_column": "will_complete_course",
  "feature_columns": [
    "study_frequency",
    "avg_session_duration",
    "quiz_performance",
    "social_engagement",
    "content_interaction"
  ],
  "model_config": {
    "algorithm": "random_forest",
    "parameters": {
      "n_estimators": 100,
      "max_depth": 10,
      "random_state": 42
    },
    "cross_validation": {
      "folds": 5,
      "stratified": true
    }
  },
  "evaluation_metrics": [
    "accuracy",
    "precision",
    "recall",
    "f1_score",
    "auc_roc"
  ]
}
```

## 6. 数据可视化服务

### 6.1 创建图表

```http
POST /api/v1/data/visualization/charts
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "title": "用户学习时长分布",
  "chart_type": "histogram",
  "data_source": {
    "type": "query",
    "query": "SELECT study_time FROM user_learning_records WHERE created_at >= '2024-01-01'"
  },
  "chart_config": {
    "x_axis": {
      "column": "study_time",
      "title": "学习时长（分钟）",
      "bins": 20
    },
    "y_axis": {
      "title": "用户数量"
    },
    "styling": {
      "color_scheme": "blue",
      "width": 800,
      "height": 600
    }
  },
  "filters": [
    {
      "column": "course_category",
      "type": "select",
      "options": ["philosophy", "health", "culture"]
    }
  ]
}
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "chart_id": "chart_1234567890abcdef",
    "title": "用户学习时长分布",
    "chart_type": "histogram",
    "chart_url": "https://cdn.taishanglaojun.com/charts/chart_1234567890abcdef.png",
    "interactive_url": "https://app.taishanglaojun.com/charts/chart_1234567890abcdef",
    "embed_code": "<iframe src='https://app.taishanglaojun.com/charts/chart_1234567890abcdef/embed' width='800' height='600'></iframe>",
    "data_summary": {
      "total_records": 10000,
      "data_range": {
        "min": 5,
        "max": 240
      }
    },
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

### 6.2 创建仪表板

```http
POST /api/v1/data/visualization/dashboards
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "title": "学习平台数据概览",
  "description": "展示平台核心学习指标的综合仪表板",
  "layout": {
    "type": "grid",
    "columns": 12,
    "rows": 8
  },
  "widgets": [
    {
      "type": "metric_card",
      "title": "总用户数",
      "position": {"x": 0, "y": 0, "w": 3, "h": 2},
      "data_source": {
        "query": "SELECT COUNT(*) as total_users FROM users"
      },
      "config": {
        "metric_format": "number",
        "trend_comparison": "last_month"
      }
    },
    {
      "type": "chart",
      "title": "每日活跃用户",
      "position": {"x": 3, "y": 0, "w": 6, "h": 4},
      "chart_id": "chart_daily_active_users"
    },
    {
      "type": "table",
      "title": "热门课程",
      "position": {"x": 0, "y": 2, "w": 6, "h": 4},
      "data_source": {
        "query": "SELECT course_title, enrollment_count FROM courses ORDER BY enrollment_count DESC LIMIT 10"
      }
    }
  ],
  "refresh_interval": 300,
  "access_control": {
    "visibility": "private",
    "shared_with": ["team_analytics", "management"]
  }
}
```

### 6.3 生成报表

```http
POST /api/v1/data/visualization/reports
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "title": "月度学习数据报表",
  "template": "monthly_learning_report",
  "parameters": {
    "month": "2024-01",
    "include_charts": true,
    "include_recommendations": true
  },
  "sections": [
    {
      "title": "用户参与度分析",
      "content": [
        {
          "type": "chart",
          "chart_id": "chart_user_engagement"
        },
        {
          "type": "text",
          "content": "本月用户参与度较上月提升15%..."
        }
      ]
    },
    {
      "title": "课程完成情况",
      "content": [
        {
          "type": "table",
          "data_source": {
            "query": "SELECT course_title, completion_rate FROM course_stats WHERE month = '2024-01'"
          }
        }
      ]
    }
  ],
  "output_format": "pdf",
  "delivery": {
    "email": ["analytics@taishanglaojun.com"],
    "schedule": "monthly"
  }
}
```

## 7. 数据治理服务

### 7.1 数据目录管理

```http
GET /api/v1/data/catalog/datasets?category=learning&status=active
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "datasets": [
      {
        "dataset_id": "dataset_1234567890abcdef",
        "name": "user_learning_records",
        "title": "用户学习记录",
        "description": "记录用户在平台上的学习活动和进度",
        "category": "learning",
        "subcategory": "user_behavior",
        "owner": {
          "team": "data_team",
          "contact": "data@taishanglaojun.com"
        },
        "schema": {
          "columns": [
            {
              "name": "user_id",
              "type": "varchar(50)",
              "nullable": false,
              "description": "用户唯一标识符",
              "tags": ["pii", "identifier"]
            },
            {
              "name": "course_id",
              "type": "varchar(50)",
              "nullable": false,
              "description": "课程唯一标识符"
            },
            {
              "name": "study_time",
              "type": "integer",
              "nullable": true,
              "description": "学习时长（秒）",
              "constraints": {
                "min": 0,
                "max": 86400
              }
            }
          ]
        },
        "data_quality": {
          "completeness": 0.98,
          "accuracy": 0.95,
          "consistency": 0.97,
          "last_checked": "2024-01-15T10:00:00Z"
        },
        "usage_stats": {
          "queries_last_30_days": 1250,
          "active_users": 15,
          "data_volume_gb": 2.5
        },
        "lineage": {
          "upstream_sources": ["user_activities", "course_enrollments"],
          "downstream_consumers": ["learning_analytics", "recommendation_engine"]
        },
        "tags": ["learning", "user_behavior", "analytics"],
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-15T08:30:00Z"
      }
    ]
  }
}
```

### 7.2 数据血缘追踪

```http
GET /api/v1/data/lineage/{dataset_id}?direction=both&depth=3
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "dataset_id": "dataset_1234567890abcdef",
    "lineage_graph": {
      "nodes": [
        {
          "id": "dataset_1234567890abcdef",
          "name": "user_learning_records",
          "type": "dataset",
          "level": 0
        },
        {
          "id": "source_user_activities",
          "name": "user_activities",
          "type": "source_table",
          "level": -1
        },
        {
          "id": "pipeline_behavior_analysis",
          "name": "behavior_analysis_pipeline",
          "type": "pipeline",
          "level": 1
        }
      ],
      "edges": [
        {
          "source": "source_user_activities",
          "target": "dataset_1234567890abcdef",
          "type": "data_flow",
          "transformation": "filter_and_aggregate"
        },
        {
          "source": "dataset_1234567890abcdef",
          "target": "pipeline_behavior_analysis",
          "type": "consumption",
          "last_accessed": "2024-01-15T09:30:00Z"
        }
      ]
    },
    "impact_analysis": {
      "upstream_changes": 2,
      "downstream_dependencies": 5,
      "critical_dependencies": ["recommendation_engine", "user_dashboard"]
    }
  }
}
```

### 7.3 数据质量监控

```http
POST /api/v1/data/quality/checks
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "dataset": "user_learning_records",
  "checks": [
    {
      "type": "completeness",
      "columns": ["user_id", "course_id"],
      "threshold": 0.99
    },
    {
      "type": "uniqueness",
      "columns": ["user_id", "course_id", "lesson_id"],
      "threshold": 1.0
    },
    {
      "type": "range",
      "column": "completion_rate",
      "min": 0,
      "max": 1
    },
    {
      "type": "pattern",
      "column": "email",
      "pattern": "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
    },
    {
      "type": "freshness",
      "column": "created_at",
      "max_age_hours": 24
    }
  ],
  "schedule": {
    "frequency": "daily",
    "time": "06:00"
  },
  "alerts": {
    "email": ["data-quality@taishanglaojun.com"],
    "slack_channel": "#data-alerts",
    "threshold": "warning"
  }
}
```

### 7.4 数据安全与隐私

```http
POST /api/v1/data/security/anonymize
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "dataset": "user_learning_records",
  "anonymization_rules": [
    {
      "column": "user_id",
      "method": "hash",
      "algorithm": "sha256",
      "salt": "learning_platform_2024"
    },
    {
      "column": "email",
      "method": "mask",
      "pattern": "***@***.***"
    },
    {
      "column": "name",
      "method": "generalize",
      "mapping": {
        "张三": "用户A",
        "李四": "用户B"
      }
    },
    {
      "column": "birth_date",
      "method": "date_shift",
      "range_days": 30
    }
  ],
  "output_dataset": "user_learning_records_anonymized",
  "preserve_relationships": true
}
```

## 8. 实时数据流处理

### 8.1 创建流处理作业

```http
POST /api/v1/data/streaming/jobs
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "job_name": "real_time_user_activity_processing",
  "description": "实时处理用户活动数据并生成洞察",
  "source": {
    "type": "kafka",
    "config": {
      "bootstrap_servers": ["kafka1:9092"],
      "topic": "user_activities",
      "consumer_group": "streaming_processor"
    }
  },
  "processing": {
    "window_type": "tumbling",
    "window_size": "5 minutes",
    "operations": [
      {
        "type": "filter",
        "condition": "activity_type IN ('page_view', 'video_watch', 'quiz_submit')"
      },
      {
        "type": "aggregate",
        "group_by": ["user_id", "activity_type"],
        "aggregations": {
          "count": "COUNT(*)",
          "total_duration": "SUM(duration)"
        }
      },
      {
        "type": "enrich",
        "lookup": {
          "source": "user_profiles",
          "key": "user_id",
          "fields": ["segment", "preferences"]
        }
      }
    ]
  },
  "sinks": [
    {
      "type": "elasticsearch",
      "config": {
        "index": "user_activity_metrics",
        "doc_type": "_doc"
      }
    },
    {
      "type": "redis",
      "config": {
        "key_pattern": "activity:{user_id}:{window_start}",
        "ttl": 3600
      }
    }
  ],
  "error_handling": {
    "strategy": "retry",
    "max_retries": 3,
    "dead_letter_queue": "failed_activities"
  }
}
```

### 8.2 监控流处理作业

```http
GET /api/v1/data/streaming/jobs/{job_id}/metrics
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "job_id": "streaming_job_1234567890abcdef",
    "status": "running",
    "metrics": {
      "throughput": {
        "records_per_second": 1250.5,
        "bytes_per_second": 156789
      },
      "latency": {
        "avg_processing_time_ms": 45.2,
        "p95_processing_time_ms": 89.1,
        "p99_processing_time_ms": 156.7
      },
      "errors": {
        "error_rate": 0.001,
        "total_errors": 12,
        "error_types": {
          "parsing_error": 8,
          "validation_error": 4
        }
      },
      "resource_usage": {
        "cpu_utilization": 0.65,
        "memory_usage_mb": 2048,
        "network_io_mbps": 15.2
      }
    },
    "checkpoints": {
      "last_checkpoint": "2024-01-15T10:25:00Z",
      "checkpoint_interval_ms": 60000,
      "checkpoint_size_mb": 125.6
    },
    "watermarks": {
      "current_watermark": "2024-01-15T10:29:30Z",
      "lag_seconds": 30
    }
  }
}
```

## 9. 数据导出服务

### 9.1 创建数据导出任务

```http
POST /api/v1/data/export
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "export_name": "monthly_learning_report_data",
  "data_source": {
    "type": "query",
    "query": "SELECT u.name, c.title, lr.completion_rate, lr.study_time FROM learning_records lr JOIN users u ON lr.user_id = u.id JOIN courses c ON lr.course_id = c.id WHERE lr.created_at >= '2024-01-01' AND lr.created_at < '2024-02-01'"
  },
  "format": "csv",
  "options": {
    "include_headers": true,
    "delimiter": ",",
    "encoding": "utf-8",
    "compression": "gzip"
  },
  "destination": {
    "type": "s3",
    "config": {
      "bucket": "data-exports",
      "path": "learning-reports/2024/01/",
      "filename": "monthly_report_2024_01.csv.gz"
    }
  },
  "schedule": {
    "type": "one_time",
    "execute_at": "2024-01-15T11:00:00Z"
  },
  "notifications": {
    "on_completion": ["data@taishanglaojun.com"],
    "on_failure": ["alerts@taishanglaojun.com"]
  }
}
```

### 9.2 获取导出任务状态

```http
GET /api/v1/data/export/{export_id}/status
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "export_id": "export_1234567890abcdef",
    "status": "completed",
    "progress": {
      "percentage": 100,
      "records_exported": 50000,
      "records_total": 50000
    },
    "timing": {
      "started_at": "2024-01-15T11:00:00Z",
      "completed_at": "2024-01-15T11:05:30Z",
      "duration_seconds": 330
    },
    "output": {
      "file_size": 5242880,
      "download_url": "https://s3.amazonaws.com/data-exports/learning-reports/2024/01/monthly_report_2024_01.csv.gz",
      "expires_at": "2024-01-22T11:05:30Z"
    },
    "metadata": {
      "format": "csv",
      "compression": "gzip",
      "encoding": "utf-8",
      "checksum": "sha256:a1b2c3d4e5f6..."
    }
  }
}
```

## 10. 数据模型

### 10.1 数据集模型

```typescript
interface Dataset {
  dataset_id: string;
  name: string;
  title: string;
  description: string;
  category: string;
  subcategory?: string;
  owner: DataOwner;
  schema: DataSchema;
  storage_info: StorageInfo;
  data_quality: DataQuality;
  usage_stats: UsageStats;
  lineage: DataLineage;
  tags: string[];
  status: 'active' | 'deprecated' | 'archived';
  created_at: string;
  updated_at: string;
}

interface DataSchema {
  columns: Column[];
  primary_key?: string[];
  foreign_keys?: ForeignKey[];
  indexes?: Index[];
}

interface Column {
  name: string;
  type: string;
  nullable: boolean;
  description?: string;
  constraints?: ColumnConstraints;
  tags?: string[];
}

interface DataQuality {
  completeness: number;
  accuracy: number;
  consistency: number;
  timeliness: number;
  last_checked: string;
  issues?: QualityIssue[];
}
```

### 10.2 数据处理管道模型

```typescript
interface DataPipeline {
  pipeline_id: string;
  name: string;
  description: string;
  version: number;
  source: DataSource;
  transformations: Transformation[];
  destination: DataDestination;
  schedule?: Schedule;
  monitoring: MonitoringConfig;
  status: 'active' | 'paused' | 'failed' | 'deprecated';
  created_at: string;
  updated_at: string;
}

interface Transformation {
  type: 'filter' | 'aggregate' | 'join' | 'enrich' | 'validate';
  config: Record<string, any>;
  order: number;
}

interface PipelineExecution {
  execution_id: string;
  pipeline_id: string;
  status: 'running' | 'completed' | 'failed' | 'cancelled';
  progress: ExecutionProgress;
  timing: ExecutionTiming;
  metrics: ExecutionMetrics;
  errors?: ExecutionError[];
}
```

## 11. 错误处理

### 11.1 数据服务特定错误

```typescript
enum DataServiceErrorCode {
  // 数据摄取错误
  INVALID_DATA_FORMAT = 'DATA1001',
  SCHEMA_VALIDATION_FAILED = 'DATA1002',
  DATA_SIZE_EXCEEDED = 'DATA1003',
  INGESTION_TIMEOUT = 'DATA1004',
  
  // 数据查询错误
  INVALID_QUERY_SYNTAX = 'DATA2001',
  QUERY_TIMEOUT = 'DATA2002',
  INSUFFICIENT_PERMISSIONS = 'DATA2003',
  DATASET_NOT_FOUND = 'DATA2004',
  
  // 数据处理错误
  PIPELINE_EXECUTION_FAILED = 'DATA3001',
  TRANSFORMATION_ERROR = 'DATA3002',
  RESOURCE_EXHAUSTED = 'DATA3003',
  DEPENDENCY_FAILURE = 'DATA3004',
  
  // 数据质量错误
  DATA_QUALITY_CHECK_FAILED = 'DATA4001',
  COMPLETENESS_THRESHOLD_NOT_MET = 'DATA4002',
  ACCURACY_THRESHOLD_NOT_MET = 'DATA4003',
  FRESHNESS_THRESHOLD_NOT_MET = 'DATA4004',
  
  // 数据安全错误
  UNAUTHORIZED_DATA_ACCESS = 'DATA5001',
  PII_EXPOSURE_DETECTED = 'DATA5002',
  ENCRYPTION_FAILED = 'DATA5003',
  AUDIT_LOG_FAILURE = 'DATA5004',
}
```

## 12. 性能优化

### 12.1 查询优化

```yaml
# 数据查询优化配置
query_optimization:
  caching:
    enabled: true
    ttl: 3600
    cache_size: "1GB"
    
  indexing:
    auto_index_creation: true
    index_usage_monitoring: true
    
  partitioning:
    strategy: "time_based"
    partition_size: "1_month"
    
  query_rewriting:
    enabled: true
    cost_based_optimization: true
```

### 12.2 存储优化

```yaml
# 数据存储优化配置
storage_optimization:
  compression:
    algorithm: "zstd"
    level: 3
    
  columnar_storage:
    enabled: true
    format: "parquet"
    
  data_lifecycle:
    hot_tier_days: 30
    warm_tier_days: 90
    cold_tier_days: 365
    archive_after_days: 1095
```

## 13. 监控与指标

### 13.1 关键指标

```yaml
# 数据服务监控指标
metrics:
  data_quality:
    - name: "data_completeness"
      description: "数据完整性"
      target: "> 95%"
      
    - name: "data_freshness"
      description: "数据新鲜度"
      target: "< 1 hour"
      
  performance:
    - name: "query_response_time"
      description: "查询响应时间"
      target: "< 2s"
      
    - name: "ingestion_throughput"
      description: "数据摄取吞吐量"
      target: "> 10MB/s"
      
  reliability:
    - name: "pipeline_success_rate"
      description: "管道成功率"
      target: "> 99%"
      
    - name: "data_availability"
      description: "数据可用性"
      target: "> 99.9%"
```

## 14. 相关文档

- [API概览文档](../api-overview.md)
- [用户服务API](./user-service-api.md)
- [AI服务API](./ai-service-api.md)
- [学习服务API](./learning-service-api.md)
- [数据治理指南](../data-governance-guide.md)
- [数据安全最佳实践](../data-security-guide.md)