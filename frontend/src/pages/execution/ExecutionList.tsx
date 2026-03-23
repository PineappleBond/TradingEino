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
  DatePicker,
  Tooltip,
} from 'antd';
import { EyeOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import type { ExecutionStatus, CronExecution } from '../../types/cronexecution';
import { getExecutionList } from '../../api/cronexecution';
import dayjs from 'dayjs';
import type { Dayjs } from 'dayjs';

const { Title } = Typography;
const { RangePicker } = DatePicker;

/**
 * 分页大小选项
 */
const pageSizeOptions = ['10', '20', '50', '100'];

/**
 * 执行状态颜色映射
 */
const statusColorMap: Record<ExecutionStatus, string> = {
  pending: 'orange',
  running: 'blue',
  success: 'green',
  failed: 'red',
  retried: 'purple',
  cancelled: 'default',
};

/**
 * 执行状态标签
 */
const statusTagMap: Record<ExecutionStatus, string> = {
  pending: '待执行',
  running: '运行中',
  success: '成功',
  failed: '失败',
  retried: '已重试',
  cancelled: '已取消',
};

/**
 * 执行记录列表页面
 */
const ExecutionList: React.FC = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [executions, setExecutions] = useState<CronExecution[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  // 筛选状态
  const [filterStatus, setFilterStatus] = useState<ExecutionStatus | undefined>();
  const [filterTimeRange, setFilterTimeRange] = useState<[Dayjs, Dayjs] | null>(null);

  /**
   * 加载执行记录列表
   */
  const loadExecutions = async () => {
    setLoading(true);
    try {
      const params: any = {
        page,
        pageSize,
        status: filterStatus,
      };

      if (filterTimeRange && filterTimeRange[0] && filterTimeRange[1]) {
        params.start_time = filterTimeRange[0].format('YYYY-MM-DD HH:mm:ss');
        params.end_time = filterTimeRange[1].format('YYYY-MM-DD HH:mm:ss');
      }

      const res = await getExecutionList(params);
      setExecutions(res.data.items);
      setTotal(res.data.page.total);
    } catch (error) {
      console.error('加载执行记录列表失败:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadExecutions();
  }, [page, pageSize, filterStatus]);

  /**
   * 表格列配置
   */
  const columns: ColumnsType<CronExecution> = [
    {
      title: 'ID',
      dataIndex: 'id',
      width: 60,
      sorter: (a, b) => a.id - b.id,
    },
    {
      title: '任务 ID',
      dataIndex: 'task_id',
      width: 80,
      render: (taskId: number) => (
        <Button type="link" size="small" onClick={() => navigate(`/task/${taskId}`)}>
          #{taskId}
        </Button>
      ),
    },
    {
      title: '计划执行时间',
      dataIndex: 'scheduled_at',
      width: 160,
      render: (time: string) => dayjs(time).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: '实际开始时间',
      dataIndex: 'started_at',
      width: 160,
      render: (time?: string) =>
        time ? dayjs(time).format('YYYY-MM-DD HH:mm:ss') : '-',
    },
    {
      title: '完成时间',
      dataIndex: 'completed_at',
      width: 160,
      render: (time?: string) =>
        time ? dayjs(time).format('YYYY-MM-DD HH:mm:ss') : '-',
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 90,
      render: (status: ExecutionStatus) => (
        <Tag color={statusColorMap[status]}>{statusTagMap[status]}</Tag>
      ),
    },
    {
      title: '重试次数',
      dataIndex: 'retry_count',
      width: 80,
      align: 'center',
    },
    {
      title: '错误信息',
      dataIndex: 'error',
      width: 200,
      ellipsis: true,
      render: (error?: string) => (
        <Tooltip title={error}>
          <span>{error || '-'}</span>
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
      render: (_: unknown, record: CronExecution) => (
        <Space size="small">
          <Tooltip title="查看详情">
            <Button
              type="link"
              size="small"
              icon={<EyeOutlined />}
              onClick={() => navigate(`/task/execution/${record.id}`)}
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
        placeholder="执行状态"
        allowClear
        style={{ width: 120 }}
        onChange={(value) => {
          setFilterStatus(value);
          setPage(1);
        }}
        options={[
          { label: '待执行', value: 'pending' },
          { label: '运行中', value: 'running' },
          { label: '成功', value: 'success' },
          { label: '失败', value: 'failed' },
          { label: '已重试', value: 'retried' },
          { label: '已取消', value: 'cancelled' },
        ]}
      />
      <RangePicker
        value={filterTimeRange}
        onChange={(dates) => {
          if (dates && dates[0] && dates[1]) {
            setFilterTimeRange([dates[0], dates[1]]);
            setPage(1);
          } else if (!dates) {
            setFilterTimeRange(null);
            setPage(1);
          }
        }}
        showTime={{ format: 'HH:mm:ss' }}
        format="YYYY-MM-DD HH:mm:ss"
      />
      <Button
        onClick={() => {
          setFilterStatus(undefined);
          setFilterTimeRange(null);
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
        <Title level={4} style={{ margin: 0, color: '#bf00ff' }}>执行记录列表</Title>
      </div>

      {filterSection}

      <Table
        rowKey="id"
        columns={columns}
        dataSource={executions}
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
        scroll={{ x: 1400 }}
      />
    </Card>
  );
};

export default ExecutionList;
