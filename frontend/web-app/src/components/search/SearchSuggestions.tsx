import React, { useState, useEffect, useRef } from 'react';
import { AutoComplete, Input, Spin, Typography } from 'antd';
import { SearchOutlined, FireOutlined } from '@ant-design/icons';
import { apiClient } from '../../services/api';
import behaviorService from '../../services/behaviorService';

const { Text } = Typography;

interface SearchSuggestionsProps {
  placeholder?: string;
  onSearch?: (value: string) => void;
  onSelect?: (value: string) => void;
  style?: React.CSSProperties;
  size?: 'small' | 'middle' | 'large';
  allowClear?: boolean;
  autoFocus?: boolean;
}

interface SuggestionOption {
  value: string;
  label: React.ReactNode;
  type: 'suggestion' | 'popular';
}

const SearchSuggestions: React.FC<SearchSuggestionsProps> = ({
  placeholder = '搜索文化智慧...',
  onSearch,
  onSelect,
  style,
  size = 'middle',
  allowClear = true,
  autoFocus = false,
}) => {
  const [options, setOptions] = useState<SuggestionOption[]>([]);
  const [loading, setLoading] = useState(false);
  const [popularSearches, setPopularSearches] = useState<string[]>([]);
  const [inputValue, setInputValue] = useState('');
  const debounceRef = useRef<NodeJS.Timeout>();

  // 加载热门搜索
  useEffect(() => {
    loadPopularSearches();
  }, []);

  const loadPopularSearches = async () => {
    try {
      const response = await apiClient.getPopularSearches(5);
      if (response.success && response.data.searches) {
        setPopularSearches(response.data.searches);
        // 如果没有输入内容，显示热门搜索
        if (!inputValue) {
          const popularOptions: SuggestionOption[] = response.data.searches.map(search => ({
            value: search,
            type: 'popular' as const,
            label: (
              <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                <FireOutlined style={{ color: '#ff4d4f', fontSize: '12px' }} />
                <Text>{search}</Text>
                <Text type="secondary" style={{ fontSize: '12px', marginLeft: 'auto' }}>
                  热门
                </Text>
              </div>
            ),
          }));
          setOptions(popularOptions);
        }
      }
    } catch (error) {
      console.warn('加载热门搜索失败:', error);
    }
  };

  // 获取搜索建议
  const fetchSuggestions = async (query: string) => {
    if (!query.trim()) {
      // 没有输入时显示热门搜索
      const popularOptions: SuggestionOption[] = popularSearches.map(search => ({
        value: search,
        type: 'popular' as const,
        label: (
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <FireOutlined style={{ color: '#ff4d4f', fontSize: '12px' }} />
            <Text>{search}</Text>
            <Text type="secondary" style={{ fontSize: '12px', marginLeft: 'auto' }}>
              热门
            </Text>
          </div>
        ),
      }));
      setOptions(popularOptions);
      return;
    }

    setLoading(true);
    try {
      const response = await apiClient.getSearchSuggestions(query, 8);
      if (response.success && response.data.suggestions) {
        const suggestionOptions: SuggestionOption[] = response.data.suggestions.map(suggestion => ({
          value: suggestion,
          type: 'suggestion' as const,
          label: (
            <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
              <SearchOutlined style={{ color: '#1890ff', fontSize: '12px' }} />
              <Text>
                {suggestion.split(new RegExp(`(${query})`, 'gi')).map((part, index) =>
                  part.toLowerCase() === query.toLowerCase() ? (
                    <span key={index} style={{ color: '#1890ff', fontWeight: 'bold' }}>
                      {part}
                    </span>
                  ) : (
                    part
                  )
                )}
              </Text>
            </div>
          ),
        }));

        // 合并搜索建议和热门搜索（如果有相关的热门搜索）
        const relevantPopular = popularSearches
          .filter(search => search.toLowerCase().includes(query.toLowerCase()))
          .slice(0, 2)
          .map(search => ({
            value: search,
            type: 'popular' as const,
            label: (
              <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                <FireOutlined style={{ color: '#ff4d4f', fontSize: '12px' }} />
                <Text>
                  {search.split(new RegExp(`(${query})`, 'gi')).map((part, index) =>
                    part.toLowerCase() === query.toLowerCase() ? (
                      <span key={index} style={{ color: '#1890ff', fontWeight: 'bold' }}>
                        {part}
                      </span>
                    ) : (
                      part
                    )
                  )}
                </Text>
                <Text type="secondary" style={{ fontSize: '12px', marginLeft: 'auto' }}>
                  热门
                </Text>
              </div>
            ),
          }));

        setOptions([...suggestionOptions, ...relevantPopular]);
      }
    } catch (error) {
      console.warn('获取搜索建议失败:', error);
      setOptions([]);
    } finally {
      setLoading(false);
    }
  };

  // 处理输入变化
  const handleSearch = (value: string) => {
    setInputValue(value);
    
    // 清除之前的防抖定时器
    if (debounceRef.current) {
      clearTimeout(debounceRef.current);
    }

    // 设置新的防抖定时器
    debounceRef.current = setTimeout(() => {
      fetchSuggestions(value);
    }, 300);
  };

  // 处理选择
  const handleSelect = async (value: string) => {
    setInputValue(value);
    
    // 记录搜索行为
    try {
      await behaviorService.recordSearch(value, 0);
    } catch (error) {
      console.warn('记录搜索行为失败:', error);
    }

    if (onSelect) {
      onSelect(value);
    }
  };

  // 处理搜索提交
  const handleSubmit = async (value: string) => {
    if (!value.trim()) return;

    // 记录搜索行为
    try {
      await behaviorService.recordSearch(value, 0);
    } catch (error) {
      console.warn('记录搜索行为失败:', error);
    }

    if (onSearch) {
      onSearch(value);
    }
  };

  // 清理防抖定时器
  useEffect(() => {
    return () => {
      if (debounceRef.current) {
        clearTimeout(debounceRef.current);
      }
    };
  }, []);

  return (
    <AutoComplete
      value={inputValue}
      options={options}
      onSearch={handleSearch}
      onSelect={handleSelect}
      style={style}
      notFoundContent={loading ? <Spin size="small" /> : null}
      dropdownMatchSelectWidth={false}
      dropdownStyle={{ minWidth: '300px' }}
    >
      <Input.Search
        placeholder={placeholder}
        size={size}
        allowClear={allowClear}
        autoFocus={autoFocus}
        onSearch={handleSubmit}
        enterButton={<SearchOutlined />}
        loading={loading}
      />
    </AutoComplete>
  );
};

export default SearchSuggestions;