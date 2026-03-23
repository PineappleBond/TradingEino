import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Table,
  Button,
  Space,
  Tag,
  Popconfirm,
  message,
  Select,
  Card,
  Typography,
  Tooltip,
} from 'antd';
import {
  PlusOutlined,
  EyeOutlined,
  EditOutlined,
  DeleteOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  CopyOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import type { TaskStatus, TaskType, CronTask } from '../../types/crontask';
import {
  getTaskList,
  deleteTask,
  enableTask,
  disableTask,
} from '../../api/crontask';
import dayjs from 'dayjs';

const { Title } = Typography;

/**
 * 分页大小选项
 */
const pageSizeOptions = ['10', '20', '50', '100'];

/**
 * 任务状态颜色映射
 */
const statusColorMap: Record<TaskStatus, string> = {
  pending: 'orange',
  running: 'blue',
  completed: 'green',
  stopped: 'default',
  failed: 'red',
};

/**
 * 任务类型标签
 */
const typeTagMap: Record<TaskType, string> = {
  once: '一次性',
  recurring: '周期性',
};

/**
 * 定时任务列表页面
 */
const TaskList: React.FC = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [tasks, setTasks] = useState<CronTask[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  // 筛选状态
  const [filterStatus, setFilterStatus] = useState<TaskStatus | undefined>();
  const [filterType, setFilterType] = useState<TaskType | undefined>();
  const [filterEnabled, setFilterEnabled] = useState<boolean | undefined>();

  /**
   * 加载任务列表
   */
  const loadTasks = async () => {
    setLoading(true);
    try {
      const res = await getTaskList({
        page,
        pageSize,
        status: filterStatus,
        type: filterType,
        enabled: filterEnabled,
      });
      setTasks(res.data.items);
      setTotal(res.data.page.total);
    } catch (error) {
      console.error('加载任务列表失败:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadTasks();
  }, [page, pageSize, filterStatus, filterType, filterEnabled]);

  /**
   * 删除任务
   */
  const handleDelete = async (id: number) => {
    try {
      await deleteTask(id);
      message.success('删除成功');
      loadTasks();
    } catch (error) {
      console.error('删除失败:', error);
    }
  };

  /**
   * 启用/禁用任务
   */
  const handleEnableDisable = async (id: number, enabled: boolean) => {
    try {
      if (enabled) {
        await enableTask(id);
        message.success('启用成功');
      } else {
        await disableTask(id);
        message.success('禁用成功');
      }
      loadTasks();
    } catch (error) {
      console.error('操作失败:', error);
    }
  };

  /**
   * 表格列配置
   */
  const columns: ColumnsType<CronTask> = [
    {
      title: 'ID',
      dataIndex: 'id',
      width: 60,
      sorter: (a, b) => a.id - b.id,
    },
    {
      title: '名称',
      dataIndex: 'name',
      width: 150,
      ellipsis: true,
    },
    {
      title: '类型',
      dataIndex: 'type',
      width: 80,
      render: (type: TaskType) => (
        <Tag color={type === 'once' ? 'purple' : 'blue'}>
          {typeTagMap[type]}
        </Tag>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 90,
      render: (status: TaskStatus) => (
        <Tag color={statusColorMap[status]}>{status}</Tag>
      ),
    },
    {
      title: 'Cron 表达式',
      dataIndex: 'spec',
      width: 120,
      ellipsis: true,
    },
    {
      title: '执行类型',
      dataIndex: 'exec_type',
      width: 100,
    },
    {
      title: 'Symbol',
      key: 'symbol',
      width: 150,
      ellipsis: true,
      render: (_: unknown, record: CronTask) => {
        try {
          const raw = JSON.parse(record.raw);
          return raw.symbol || '-';
        } catch {
          return '-';
        }
      },
    },
    {
      title: '启用',
      dataIndex: 'enabled',
      width: 60,
      render: (enabled: boolean) => (
        <Tag color={enabled ? 'green' : 'default'}>
          {enabled ? '是' : '否'}
        </Tag>
      ),
    },
    {
      title: '下次执行时间',
      dataIndex: 'next_execution_at',
      width: 160,
      render: (time?: string) => (time ? dayjs(time).format('YYYY-MM-DD HH:mm:ss') : '-'),
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
      width: 250,
      fixed: 'right',
      render: (_: unknown, record: CronTask) => (
        <Space size="small" wrap>
          <Tooltip title="查看详情">
            <Button
              type="link"
              size="small"
              icon={<EyeOutlined />}
              onClick={() => navigate(`/task/${record.id}`)}
            />
          </Tooltip>
          <Tooltip title="编辑">
            <Button
              type="link"
              size="small"
              icon={<EditOutlined />}
              onClick={() => navigate(`/task/${record.id}/edit`)}
            />
          </Tooltip>
          <Tooltip title="拷贝">
            <Button
              type="link"
              size="small"
              icon={<CopyOutlined />}
              onClick={() => navigate(`/task/create?copyFrom=${record.id}`)}
            />
          </Tooltip>
          <Tooltip title={record.enabled ? '禁用' : '启用'}>
            <Button
              type="link"
              size="small"
              icon={record.enabled ? <CloseCircleOutlined /> : <CheckCircleOutlined />}
              onClick={() => handleEnableDisable(record.id, !record.enabled)}
            />
          </Tooltip>
          <Popconfirm
            title="确定要删除此任务吗？"
            onConfirm={() => handleDelete(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Button type="link" size="small" danger icon={<DeleteOutlined />} />
          </Popconfirm>
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
        placeholder="任务状态"
        allowClear
        style={{ width: 120 }}
        onChange={(value) => {
          setFilterStatus(value);
          setPage(1);
        }}
        options={[
          { label: '待执行', value: 'pending' },
          { label: '运行中', value: 'running' },
          { label: '已完成', value: 'completed' },
          { label: '已停止', value: 'stopped' },
          { label: '失败', value: 'failed' },
        ]}
      />
      <Select
        placeholder="任务类型"
        allowClear
        style={{ width: 100 }}
        onChange={(value) => {
          setFilterType(value);
          setPage(1);
        }}
        options={[
          { label: '一次性', value: 'once' },
          { label: '周期性', value: 'recurring' },
        ]}
      />
      <Select
        placeholder="启用状态"
        allowClear
        style={{ width: 100 }}
        onChange={(value) => {
          setFilterEnabled(value);
          setPage(1);
        }}
        options={[
          { label: '启用', value: true },
          { label: '禁用', value: false },
        ]}
      />
    </Space>
  );

  return (
    <Card className="cyber-card" style={{ margin: 16 }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0, color: '#bf00ff' }}>定时任务列表</Title>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => navigate('/task/create')}
        >
          新建任务
        </Button>
      </div>

      {filterSection}

      <Table
        rowKey="id"
        columns={columns}
        dataSource={tasks}
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
        scroll={{ x: 1500 }}
      />
    </Card>
  );
};

export default TaskList;
