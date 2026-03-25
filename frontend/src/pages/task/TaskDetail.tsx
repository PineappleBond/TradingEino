import { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
  Card,
  Descriptions,
  Button,
  Space,
  Tag,
  message,
  Modal,
  Form,
  DatePicker,
  Typography,
} from 'antd';
import {
  ArrowLeftOutlined,
  EditOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
} from '@ant-design/icons';
import type { CronTask, TaskStatus } from '../../types/crontask';
import { getTaskDetail, startTask, enableTask, disableTask } from '../../api/crontask';
import dayjs from 'dayjs';

const { Title, Paragraph } = Typography;

/**
 * 定时任务详情页面
 */
const TaskDetail: React.FC = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const [task, setTask] = useState<CronTask | null>(null);
  const [startModalOpen, setStartModalOpen] = useState(false);
  const [startForm] = Form.useForm();

  /**
   * 加载任务详情
   */
  const loadTaskDetail = async () => {
    if (!id) return;
    try {
      const res = await getTaskDetail(parseInt(id, 10));
      setTask(res.data);
    } catch (error) {
      console.error('加载任务详情失败:', error);
    }
  };

  useEffect(() => {
    loadTaskDetail();
  }, [id]);

  /**
   * 启动任务
   */
  const handleStart = async () => {
    try {
      const values = await startForm.validateFields();
      if (!id) return;

      await startTask(parseInt(id, 10), {
        next_execution_time: values.next_execution_time,
      });
      message.success('任务启动成功');
      setStartModalOpen(false);
      loadTaskDetail();
    } catch (error) {
      console.error('启动任务失败:', error);
    }
  };

  /**
   * 启用/禁用任务
   */
  const handleEnableDisable = async (enabled: boolean) => {
    try {
      if (!id) return;
      if (enabled) {
        await enableTask(parseInt(id, 10));
        message.success('启用成功');
      } else {
        await disableTask(parseInt(id, 10));
        message.success('禁用成功');
      }
      loadTaskDetail();
    } catch (error) {
      console.error('操作失败:', error);
    }
  };

  if (!task) {
    return (
      <Card style={{ margin: 16 }}>
        <Paragraph>加载中...</Paragraph>
      </Card>
    );
  }

  /**
   * 解析 raw JSON
   */
  let rawSymbol = '-';
  try {
    const raw = JSON.parse(task.raw);
    rawSymbol = raw.symbol || '-';
  } catch {
    rawSymbol = '无效 JSON';
  }

  return (
    <Card
      className="cyber-card"
      style={{ margin: 16 }}
      title={
        <Space>
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate('/task')}
          >
            返回
          </Button>
          <Title level={4} style={{ margin: 0 }}>任务详情</Title>
        </Space>
      }
      extra={
        <Space>
          <Button
            icon={<EditOutlined />}
            onClick={() => navigate(`/task/${id}/edit`)}
          >
            编辑
          </Button>
          {task.enabled ? (
            <Button
              icon={<CloseCircleOutlined />}
              onClick={() => handleEnableDisable(false)}
            >
              禁用
            </Button>
          ) : (
            <Button
              icon={<CheckCircleOutlined />}
              onClick={() => handleEnableDisable(true)}
            >
              启用
            </Button>
          )}
        </Space>
      }
    >
      <Descriptions bordered column={2}>
        <Descriptions.Item label="ID">{task.id}</Descriptions.Item>
        <Descriptions.Item label="名称">{task.name}</Descriptions.Item>
        <Descriptions.Item label="类型">
          <Tag color={task.type === 'once' ? 'purple' : 'blue'}>
            {task.type === 'once' ? '一次性' : '周期性'}
          </Tag>
        </Descriptions.Item>
        <Descriptions.Item label="状态">
          <Tag color={getStatusColor(task.status)}>{task.status}</Tag>
        </Descriptions.Item>
        <Descriptions.Item label="Cron 表达式" span={2}>
          <Paragraph code style={{ margin: 0 }}>{task.spec}</Paragraph>
        </Descriptions.Item>
        <Descriptions.Item label="执行类型">{task.exec_type}</Descriptions.Item>
        <Descriptions.Item label="Symbol">{rawSymbol}</Descriptions.Item>
        <Descriptions.Item label="启用">
          <Tag color={task.enabled ? 'green' : 'default'}>
            {task.enabled ? '是' : '否'}
          </Tag>
        </Descriptions.Item>
        <Descriptions.Item label="最大重试次数">{task.max_retries}</Descriptions.Item>
        <Descriptions.Item label="超时时间 (秒)">{task.timeout_seconds}</Descriptions.Item>
        <Descriptions.Item label="总执行次数">{task.total_executions}</Descriptions.Item>
        <Descriptions.Item label="有效期">
          {dayjs(task.valid_from).format('YYYY-MM-DD HH:mm:ss')} ~ {dayjs(task.valid_until).format('YYYY-MM-DD HH:mm:ss')}
        </Descriptions.Item>
        <Descriptions.Item label="上次执行时间">
          {task.last_executed_at
            ? dayjs(task.last_executed_at).format('YYYY-MM-DD HH:mm:ss')
            : '-'}
        </Descriptions.Item>
        <Descriptions.Item label="下次执行时间">
          {task.next_execution_at
            ? dayjs(task.next_execution_at).format('YYYY-MM-DD HH:mm:ss')
            : '-'}
        </Descriptions.Item>
        <Descriptions.Item label="创建时间">
          {dayjs(task.created_at).format('YYYY-MM-DD HH:mm:ss')}
        </Descriptions.Item>
        <Descriptions.Item label="更新时间">
          {dayjs(task.updated_at).format('YYYY-MM-DD HH:mm:ss')}
        </Descriptions.Item>
      </Descriptions>

      {/* 启动任务弹窗 */}
      <Modal
        title="启动任务"
        open={startModalOpen}
        onOk={handleStart}
        onCancel={() => setStartModalOpen(false)}
        okText="确定"
        cancelText="取消"
      >
        <Form form={startForm} layout="vertical">
          <Form.Item
            name="next_execution_time"
            label="下次执行时间"
            rules={[{ required: true, message: '请选择下次执行时间' }]}
          >
            <DatePicker
              showTime
              format="YYYY-MM-DD HH:mm:ss"
              style={{ width: '100%' }}
              placeholder="选择执行时间"
            />
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  );
};

/**
 * 获取状态颜色
 */
function getStatusColor(status: TaskStatus): string {
  const colorMap: Record<TaskStatus, string> = {
    pending: 'orange',
    running: 'blue',
    completed: 'green',
    stopped: 'default',
    failed: 'red',
  };
  return colorMap[status];
}

export default TaskDetail;
