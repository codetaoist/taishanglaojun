import { apiClient } from './api';

// 菜单数据类型定义
export interface MenuPermission {
  id: string;
  name: string;
  code: string;
  description?: string;
}

export interface MenuItem {
  id: string;
  name: string;
  path: string;
  icon?: string;
  component?: string;
  parentId?: string;
  sort: number;
  status: 'active' | 'inactive';
  type: 'menu' | 'button' | 'api';
  permissions: MenuPermission[];
  children?: MenuItem[];
  createdAt: string;
  updatedAt: string;
}

export interface CreateMenuRequest {
  name: string;
  path: string;
  icon?: string;
  component?: string;
  parentId?: string;
  sort: number;
  status: 'active' | 'inactive';
  type: 'menu' | 'button' | 'api';
  permissions: string[];
}

export interface UpdateMenuRequest extends Partial<CreateMenuRequest> {
  id: string;
}

export interface MenuListResponse {
  items: MenuItem[];
  total: number;
  page: number;
  pageSize: number;
}

export interface MenuTreeResponse {
  items: MenuItem[];
}

// 菜单管理API服务类
export class MenuService {
  private baseUrl = '/menus';

  // 将前端字段映射为后端字段（创建）
  private toBackendCreatePayload(data: CreateMenuRequest) {
    return {
      name: data.name,
      title: data.name,
      path: data.path,
      icon: data.icon,
      parent_id: data.parentId ?? null,
      sort: data.sort,
      is_enabled: data.status === 'active',
      is_visible: true,
    };
  }

  // 将前端字段映射为后端字段（更新）
  private toBackendUpdatePayload(data: Partial<CreateMenuRequest>) {
    const updates: any = {};
    if (data.name !== undefined) {
      updates.name = data.name;
      updates.title = data.name;
    }
    if (data.path !== undefined) updates.path = data.path;
    if (data.icon !== undefined) updates.icon = data.icon;
    if (data.parentId !== undefined) updates.parent_id = data.parentId ?? null;
    if (data.sort !== undefined) updates.sort = data.sort;
    if (data.status !== undefined) updates.is_enabled = data.status === 'active';
    return updates;
  }

  // 数据转换：后端格式转前端格式
  private transformBackendMenuItem(backendItem: any): MenuItem {
    return {
      id: backendItem.id,
      name: backendItem.name,
      path: backendItem.path,
      icon: backendItem.icon,
      component: backendItem.component,
      parentId: backendItem.parent_id,
      sort: backendItem.sort,
      status: backendItem.is_enabled ? 'active' : 'inactive',
      type: this.mapBackendTypeToFrontend(backendItem),
      permissions: [], // 暂时为空，后续可以根据需要添加权限映射
      children: backendItem.children ? backendItem.children.map((child: any) => this.transformBackendMenuItem(child)) : undefined,
      createdAt: backendItem.created_at,
      updatedAt: backendItem.updated_at
    };
  }

  // 映射后端类型到前端类型
  private mapBackendTypeToFrontend(backendItem: any): 'menu' | 'button' | 'api' {
    // 根据后端数据特征判断类型
    // 若存在子节点，视为目录型菜单（即使路径为空）
    if (Array.isArray(backendItem.children) && backendItem.children.length > 0) {
      return 'menu';
    }
    if (backendItem.path && backendItem.path.startsWith('/')) {
      return 'menu';
    }
    if (backendItem.component) {
      return 'button';
    }
    return 'api';
  }

  // 获取菜单列表（分页）
  async getMenuList(params?: {
    page?: number;
    pageSize?: number;
    name?: string;
    status?: string;
    type?: string;
    parentId?: string;
  }): Promise<MenuListResponse> {
    try {
      const response = await apiClient.get<any>(`${this.baseUrl}`, { params });
      
      // 处理后端返回的数据格式: {code: 200, data: [...], message: "success"}
      const backendResponse = response.data;
      
      // 从后端响应中提取菜单数据
      let menuData = [];
      if (backendResponse && backendResponse.data && Array.isArray(backendResponse.data)) {
        menuData = backendResponse.data;
      } else if (Array.isArray(backendResponse)) {
        // 兼容直接返回数组的情况
        menuData = backendResponse;
      }
      
      const items = menuData.map((item: any) => this.transformBackendMenuItem(item));

      return {
        items,
        total: backendResponse.total || items.length,
        page: backendResponse.page || params?.page || 1,
        pageSize: backendResponse.pageSize || params?.pageSize || 10
      };
    } catch (error) {
      console.error('获取菜单列表失败:', error);
      throw error;
    }
  }

  // 获取菜单树形结构
  async getMenuTree(): Promise<MenuTreeResponse> {
    try {
      const response = await apiClient.get<any>(`${this.baseUrl}/tree`);
      
      // 处理后端返回的数据格式: {code: 200, data: [...], message: "success"}
      const backendResponse = response.data;
      console.log('🔍 MenuService.getMenuTree - 后端响应:', backendResponse);
      
      // 从后端响应中提取菜单数据
      let menuData = [];
      if (backendResponse && backendResponse.data && Array.isArray(backendResponse.data)) {
        menuData = backendResponse.data;
      } else if (Array.isArray(backendResponse)) {
        // 兼容直接返回数组的情况
        menuData = backendResponse;
      }
      
      console.log('🔍 MenuService.getMenuTree - 提取的菜单数据:', menuData);
      
      const items = menuData.map((item: any) => this.transformBackendMenuItem(item));
      
      console.log('🔍 MenuService.getMenuTree - 转换后的菜单项:', items);

      return {
        items
      };
    } catch (error) {
      console.error('获取菜单树失败:', error);
      throw error;
    }
  }

  // 获取单个菜单详情
  async getMenuById(id: string): Promise<MenuItem> {
    try {
      const response = await apiClient.get<any>(`${this.baseUrl}/${id}`);
      
      // 处理后端返回的数据格式
      const backendData = response.data;
      return this.transformBackendMenuItem(backendData.data || backendData);
    } catch (error) {
      console.error('获取菜单详情失败:', error);
      throw error;
    }
  }

  // 创建菜单
  async createMenu(data: CreateMenuRequest): Promise<MenuItem> {
    try {
      const payload = this.toBackendCreatePayload(data);
      const response = await apiClient.post<any>('/admin/menus', payload);
      const backendData = response.data;
      const created = backendData?.data || backendData;
      return this.transformBackendMenuItem(created);
    } catch (error) {
      console.error('创建菜单失败:', error);
      throw error;
    }
  }

  // 更新菜单
  async updateMenu(data: UpdateMenuRequest): Promise<MenuItem> {
    try {
      const payload = this.toBackendUpdatePayload(data);
      await apiClient.put<any>(`/admin/menus/${data.id}`, payload);
      // 更新成功后读取最新详情
      return await this.getMenuById(data.id);
    } catch (error) {
      console.error('更新菜单失败:', error);
      throw error;
    }
  }

  // 删除菜单
  async deleteMenu(id: string): Promise<void> {
    try {
      await apiClient.delete(`/admin/menus/${id}`);
    } catch (error) {
      console.error('删除菜单失败:', error);
      throw error;
    }
  }

  // 批量删除菜单
  async batchDeleteMenus(ids: string[]): Promise<void> {
    try {
      await apiClient.delete('/admin/menus/batch', { data: { ids } });
    } catch (error) {
      console.error('批量删除菜单失败:', error);
      throw error;
    }
  }

  // 更新菜单状态
  async updateMenuStatus(id: string, status: 'active' | 'inactive'): Promise<void> {
    try {
      await apiClient.patch(`/admin/menus/${id}/status`, { status });
    } catch (error) {
      console.error('更新菜单状态失败:', error);
      throw error;
    }
  }

  // 获取可用权限列表
  async getAvailablePermissions(): Promise<MenuPermission[]> {
    try {
      console.log('🔄 MenuService: 开始请求权限列表...');
      console.log('📡 请求URL: /permissions');
      console.log('📋 请求参数:', { page_size: 1000 });
      
      // 统一兼容后端多种返回格式，并映射为前端 MenuPermission[]
      const response = await apiClient.get<any>('/permissions', {
        params: { page_size: 1000 }
      });
      
      console.log('📥 API响应状态:', response.status);
      console.log('📄 API响应数据:', response.data);
      
      const raw = response.data;
      const list = Array.isArray(raw)
        ? raw
        : Array.isArray(raw?.permissions)
          ? raw.permissions
          : Array.isArray(raw?.data?.permissions)
            ? raw.data.permissions
            : Array.isArray(raw?.data)
              ? raw.data
              : Array.isArray(raw?.items)
                ? raw.items
                : [];

      console.log('🔍 解析后的权限列表:', list);
      console.log('📊 权限列表长度:', list?.length);

      const mappedPermissions = (list || []).map((p: any) => ({
        id: p.id || p.permission_id || p.ID,
        name: p.name || p.code || `${p.resource ?? ''}:${p.action ?? ''}`.replace(/^:/, ''),
        code: p.code || (p.resource && p.action ? `${p.resource}:${p.action}` : (p.name || '')),
        description: p.description || ''
      })) as MenuPermission[];
      
      console.log('✅ 映射后的权限数据:', mappedPermissions);
      return mappedPermissions;
    } catch (error) {
      console.error('❌ MenuService: 获取权限列表失败:', error);
      console.error('🔍 错误详情:', {
        message: error instanceof Error ? error.message : '未知错误',
        response: (error as any)?.response?.data,
        status: (error as any)?.response?.status,
        statusText: (error as any)?.response?.statusText,
        config: (error as any)?.config
      });
      throw error;
    }
  }


}

// 导出服务实例
export const menuService = new MenuService();