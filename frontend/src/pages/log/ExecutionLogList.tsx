import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Table,
  Button,
  Space,
  Tag,
  Card,
  Typography,
  Select,
  Tooltip,
} from 'antd';
import { EyeOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import type { LogLevel, CronExecutionLog } from '../../types/cronexecutionlog';
import { getExecutionLogList } from '../../api/cronexecutionlog';
import dayjs from 'dayjs';

const { Title } = Typography;

/**
 * 分页大小选项
 */
const pageSizeOptions = ['10', '20', '50', '100'];

/**
 * 日志级别颜色映射
 */
const logLevelColorMap: Record<string, string> = {
  info: 'blue',
  warn: 'orange',
  error: 'red',
  debug: 'purple',
};

/**
 * 执行日志列表页面
 */
const ExecutionLogList: React.FC = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [logs, setLogs] = useState<CronExecutionLog[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  // 筛选条件
  const [filterLevel, setFilterLevel] = useState<LogLevel | undefined>();

  /**
   * 加载执行日志列表
   */
  const loadLogs = async () => {
    setLoading(true);
    try {
      const res = await getExecutionLogList({
        page,
        pageSize,
        level: filterLevel,
      });
      setLogs(res.data.items);
      setTotal(res.data.page.total);
    } catch (error) {
      console.error('加载执行日志列表失败:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadLogs();
  }, [page, pageSize, filterLevel]);

  /**
   * 表格列配置
   */
  const columns: ColumnsType<CronExecutionLog> = [
    {
      title: 'ID',
      dataIndex: 'id',
      width: 60,
      sorter: (a, b) => a.id - b.id,
    },
    {
      title: '执行 ID',
      dataIndex: 'execution_id',
      width: 80,
      render: (executionId: number) => (
        <Button type="link" size="small" onClick={() => navigate(`/task/execution/${executionId}`)}>
          #{executionId}
        </Button>
      ),
    },
    {
      title: '级别',
      dataIndex: 'level',
      width: 80,
      render: (level: LogLevel) => (
        <Tag color={logLevelColorMap[level.toLowerCase()] || 'default'}>
          {level.toUpperCase()}
        </Tag>
      ),
    },
    {
      title: '来源',
      dataIndex: 'from',
      width: 120,
      ellipsis: true,
    },
    {
      title: '消息',
      dataIndex: 'message',
      ellipsis: true,
      render: (msg: string) => (
        <Tooltip title={msg}>
          <span>{msg}</span>
        </Tooltip>
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      width: 160,
      render: (time: string) => dayjs(time).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: '操作',
      key: 'action',
      width: 80,
      fixed: 'right',
      render: (_: unknown, record: CronExecutionLog) => (
        <Space size="small">
          <Tooltip title="查看详情">
            <Button
              type="link"
              size="small"
              icon={<EyeOutlined />}
              onClick={() => navigate(`/log/execution/${record.id}`)}
            />
          </Tooltip>
        </Space>
      ),
    },
  ];

  /**
   * 筛选器
   */
  const filterSection = (
    <Space wrap style={{ marginBottom: 16 }}>
      <Select
        placeholder="日志级别"
        allowClear
        style={{ width: 120 }}
        onChange={(value) => {
          setFilterLevel(value);
          setPage(1);
        }}
        options={[
          { label: 'INFO', value: 'info' },
          { label: 'WARN', value: 'warn' },
          { label: 'ERROR', value: 'error' },
          { label: 'DEBUG', value: 'debug' },
        ]}
      />
      <Button
        onClick={() => {
          setFilterLevel(undefined);
          setPage(1);
        }}
      >
        重置筛选
      </Button>
    </Space>
  );

  return (
    <Card className="cyber-card" style={{ margin: 16 }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0, color: '#bf00ff' }}>执行日志列表</Title>
      </div>

      {filterSection}

      <Table
        rowKey="id"
        columns={columns}
        dataSource={logs}
        loading={loading}
        pagination={{
          current: page,
          pageSize,
          total,
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (t) => `共 ${t} 条`,
          pageSizeOptions,
          onChange: (p, s) => {
            setPage(p);
            setPageSize(s);
          },
        }}
        scroll={{ x: 1000 }}
      />
    </Card>
  );
};

export default ExecutionLogList;
