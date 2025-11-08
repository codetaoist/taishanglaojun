from pydantic import BaseModel, Field
from typing import List, Dict, Any, Optional
from enum import Enum


class IndexType(str, Enum):
    """向量索引类型"""
    FLAT = "FLAT"
    IVF_FLAT = "IVF_FLAT"
    IVF_SQ8 = "IVF_SQ8"
    HNSW = "HNSW"


class MetricType(str, Enum):
    """距离度量类型"""
    L2 = "L2"
    IP = "IP"
    COSINE = "COSINE"
    HAMMING = "HAMMING"
    JACCARD = "JACCARD"
    TANIMOTO = "TANIMOTO"
    SUBSTRUCTURE = "SUBSTRUCTURE"
    SUPERSTRUCTURE = "SUPERSTRUCTURE"


class VectorModel(BaseModel):
    """向量数据模型"""
    id: str = Field(..., description="向量ID")
    vector: List[float] = Field(..., description="向量数据")
    metadata: Optional[Dict[str, Any]] = Field(None, description="元数据")


class SearchResultModel(BaseModel):
    """向量搜索结果模型"""
    id: str = Field(..., description="向量ID")
    score: float = Field(..., description="相似度分数")
    vector: Optional[List[float]] = Field(None, description="向量数据")
    metadata: Optional[Dict[str, Any]] = Field(None, description="元数据")


class SearchOptionsModel(BaseModel):
    """向量搜索选项"""
    topk: int = Field(10, description="返回结果数量")
    include_vector: bool = Field(False, description="是否包含向量数据")
    filter: Optional[Dict[str, Any]] = Field(None, description="过滤条件")


class IndexParamsModel(BaseModel):
    """索引参数"""
    index_type: IndexType = Field(..., description="索引类型")
    metric_type: MetricType = Field(..., description="距离度量类型")
    params: Dict[str, Any] = Field(default_factory=dict, description="索引参数")


class CollectionInfoModel(BaseModel):
    """集合信息"""
    name: str = Field(..., description="集合名称")
    description: str = Field("", description="集合描述")
    vector_dim: int = Field(..., description="向量维度")
    index_params: IndexParamsModel = Field(..., description="索引参数")
    created_at: int = Field(..., description="创建时间戳")
    updated_at: int = Field(..., description="更新时间戳")


class CollectionStatsModel(BaseModel):
    """集合统计信息"""
    name: str = Field(..., description="集合名称")
    vector_count: int = Field(..., description="向量数量")
    size_in_bytes: int = Field(..., description="大小（字节）")


class CreateCollectionRequestModel(BaseModel):
    """创建集合请求"""
    name: str = Field(..., description="集合名称")
    description: str = Field("", description="集合描述")
    vector_dim: int = Field(..., description="向量维度")
    index_params: IndexParamsModel = Field(..., description="索引参数")


class InsertRequestModel(BaseModel):
    """插入向量请求"""
    collection_name: str = Field(..., description="集合名称")
    vectors: List[VectorModel] = Field(..., description="向量列表")


class SearchRequestModel(BaseModel):
    """向量搜索请求"""
    collection_name: str = Field(..., description="集合名称")
    query_vector: List[float] = Field(..., description="查询向量")
    options: SearchOptionsModel = Field(default_factory=SearchOptionsModel, description="搜索选项")