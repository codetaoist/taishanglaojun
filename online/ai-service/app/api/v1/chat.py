#!/usr/bin/env python3
"""
码道 AI智能体服务聊天API

处理自然语言编程指令和AI对话。
"""

import time
from typing import List, Optional, Dict, Any
from fastapi import APIRouter, Depends, HTTPException, status
from pydantic import BaseModel, Field
from loguru import logger

from ...core.auth import get_current_active_user, check_rate_limit
from ...core.monitoring import monitor_requests, monitor_ai_calls
from ...services.ai_service import AIService
from ...services.context_service import ContextService


router = APIRouter()


class ChatMessage(BaseModel):
    """聊天消息模型"""
    role: str = Field(..., description="消息角色: user, assistant, system")
    content: str = Field(..., description="消息内容")
    timestamp: Optional[float] = Field(default_factory=time.time, description="时间戳")


class ChatRequest(BaseModel):
    """聊天请求模型"""
    message: str = Field(..., description="用户消息", min_length=1, max_length=10000)
    project_id: Optional[str] = Field(None, description="项目ID")
    context: Optional[Dict[str, Any]] = Field(default_factory=dict, description="上下文信息")
    model: Optional[str] = Field(None, description="指定使用的AI模型")
    stream: bool = Field(False, description="是否流式响应")
    include_plan: bool = Field(True, description="是否包含执行计划")


class ExecutionStep(BaseModel):
    """执行步骤模型"""
    step_id: int = Field(..., description="步骤ID")
    description: str = Field(..., description="步骤描述")
    status: str = Field("pending", description="步骤状态: pending, running, completed, failed")
    result: Optional[str] = Field(None, description="执行结果")
    duration: Optional[float] = Field(None, description="执行时长（秒）")


class ChatResponse(BaseModel):
    """聊天响应模型"""
    message: str = Field(..., description="AI回复消息")
    conversation_id: str = Field(..., description="对话ID")
    execution_plan: Optional[List[ExecutionStep]] = Field(None, description="执行计划")
    context_used: List[str] = Field(default_factory=list, description="使用的上下文")
    model_used: str = Field(..., description="使用的AI模型")
    tokens_used: Optional[Dict[str, int]] = Field(None, description="Token使用情况")
    timestamp: float = Field(default_factory=time.time, description="响应时间戳")


class PlanExecutionRequest(BaseModel):
    """计划执行请求模型"""
    conversation_id: str = Field(..., description="对话ID")
    confirm_execution: bool = Field(..., description="确认执行")


class PlanExecutionResponse(BaseModel):
    """计划执行响应模型"""
    execution_id: str = Field(..., description="执行ID")
    status: str = Field(..., description="执行状态")
    steps: List[ExecutionStep] = Field(..., description="执行步骤")
    results: List[str] = Field(default_factory=list, description="执行结果")
    timestamp: float = Field(default_factory=time.time, description="执行时间戳")


@router.post("/", response_model=ChatResponse)
@monitor_requests
async def chat(
    request: ChatRequest,
    current_user: Dict[str, Any] = Depends(check_rate_limit)
) -> ChatResponse:
    """
    AI聊天接口
    
    处理用户的自然语言编程指令，返回AI回复和执行计划。
    """
    try:
        logger.info(f"用户 {current_user['user_id']} 发起聊天请求")
        
        # 获取AI服务实例
        ai_service = AIService()
        context_service = ContextService()
        
        # 构建上下文
        context_data = await context_service.build_context(
            user_id=current_user["user_id"],
            project_id=request.project_id,
            additional_context=request.context
        )
        
        # 调用AI服务
        response = await ai_service.process_instruction(
            instruction=request.message,
            user_id=current_user["user_id"],
            project_id=request.project_id,
            context=context_data,
            model=request.model,
            include_plan=request.include_plan
        )
        
        return ChatResponse(
            message=response["message"],
            conversation_id=response["conversation_id"],
            execution_plan=response.get("execution_plan"),
            context_used=response.get("context_used", []),
            model_used=response["model_used"],
            tokens_used=response.get("tokens_used")
        )
        
    except Exception as e:
        logger.error(f"聊天请求处理失败: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"AI服务暂时不可用: {str(e)}"
        )


@router.post("/execute", response_model=PlanExecutionResponse)
@monitor_requests
async def execute_plan(
    request: PlanExecutionRequest,
    current_user: Dict[str, Any] = Depends(check_rate_limit)
) -> PlanExecutionResponse:
    """
    执行AI生成的计划
    
    用户确认后执行AI生成的操作计划。
    """
    try:
        logger.info(f"用户 {current_user['user_id']} 请求执行计划 {request.conversation_id}")
        
        if not request.confirm_execution:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="必须确认执行计划"
            )
        
        # 获取AI服务实例
        ai_service = AIService()
        
        # 执行计划
        execution_result = await ai_service.execute_plan(
            conversation_id=request.conversation_id,
            user_id=current_user["user_id"]
        )
        
        return PlanExecutionResponse(
            execution_id=execution_result["execution_id"],
            status=execution_result["status"],
            steps=execution_result["steps"],
            results=execution_result.get("results", [])
        )
        
    except Exception as e:
        logger.error(f"计划执行失败: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"计划执行失败: {str(e)}"
        )


@router.get("/history")
@monitor_requests
async def get_chat_history(
    project_id: Optional[str] = None,
    limit: int = 20,
    offset: int = 0,
    current_user: Dict[str, Any] = Depends(get_current_active_user)
) -> Dict[str, Any]:
    """
    获取聊天历史
    
    返回用户的聊天历史记录。
    """
    try:
        logger.info(f"用户 {current_user['user_id']} 请求聊天历史")
        
        # 模拟聊天历史数据
        history = [
            {
                "conversation_id": "conv_123",
                "timestamp": time.time() - 3600,
                "user_message": "用Go写一个HTTP服务器",
                "ai_response": "我来帮您创建一个Go语言的HTTP服务器...",
                "project_id": project_id,
                "model_used": "gpt-4",
                "status": "completed"
            },
            {
                "conversation_id": "conv_124",
                "timestamp": time.time() - 1800,
                "user_message": "解释这个函数的作用",
                "ai_response": "这个函数的主要作用是...",
                "project_id": project_id,
                "model_used": "gpt-4",
                "status": "completed"
            }
        ]
        
        # 应用分页
        total = len(history)
        paginated_history = history[offset:offset + limit]
        
        return {
            "history": paginated_history,
            "total": total,
            "limit": limit,
            "offset": offset,
            "has_more": offset + limit < total
        }
        
    except Exception as e:
        logger.error(f"获取聊天历史失败: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"获取聊天历史失败: {str(e)}"
        )


@router.get("/conversation/{conversation_id}")
@monitor_requests
async def get_conversation(
    conversation_id: str,
    current_user: Dict[str, Any] = Depends(get_current_active_user)
) -> Dict[str, Any]:
    """
    获取特定对话详情
    
    返回指定对话的完整信息。
    """
    try:
        logger.info(f"用户 {current_user['user_id']} 请求对话详情 {conversation_id}")
        
        # 模拟对话数据
        conversation = {
            "conversation_id": conversation_id,
            "user_id": current_user["user_id"],
            "project_id": "proj_1234567890",
            "created_at": time.time() - 3600,
            "updated_at": time.time() - 3500,
            "messages": [
                {
                    "role": "user",
                    "content": "用Go写一个HTTP服务器",
                    "timestamp": time.time() - 3600
                },
                {
                    "role": "assistant",
                    "content": "我来帮您创建一个Go语言的HTTP服务器。首先，我需要了解您的具体需求...",
                    "timestamp": time.time() - 3590
                }
            ],
            "execution_plan": [
                {
                    "step_id": 1,
                    "description": "创建main.go文件",
                    "status": "completed",
                    "result": "文件已创建",
                    "duration": 0.5
                },
                {
                    "step_id": 2,
                    "description": "添加HTTP路由",
                    "status": "completed",
                    "result": "路由已添加",
                    "duration": 0.3
                }
            ],
            "model_used": "gpt-4",
            "tokens_used": {
                "prompt_tokens": 150,
                "completion_tokens": 300,
                "total_tokens": 450
            },
            "status": "completed"
        }
        
        return conversation
        
    except Exception as e:
        logger.error(f"获取对话详情失败: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"获取对话详情失败: {str(e)}"
        )


@router.delete("/conversation/{conversation_id}")
@monitor_requests
async def delete_conversation(
    conversation_id: str,
    current_user: Dict[str, Any] = Depends(get_current_active_user)
) -> Dict[str, str]:
    """
    删除对话
    
    删除指定的对话记录。
    """
    try:
        logger.info(f"用户 {current_user['user_id']} 请求删除对话 {conversation_id}")
        
        # 这里应该实现实际的删除逻辑
        # 检查用户权限，删除数据库记录等
        
        return {
            "message": "对话已删除",
            "conversation_id": conversation_id
        }
        
    except Exception as e:
        logger.error(f"删除对话失败: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"删除对话失败: {str(e)}"
        )