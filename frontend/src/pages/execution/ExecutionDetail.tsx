import { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
  Card,
  Descriptions,
  Button,
  Space,
  Tag,
  Typography,
  Divider,
} from 'antd';
import { ArrowLeftOutlined, ClockCircleOutlined } from '@ant-design/icons';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeRaw from 'rehype-raw';
import type { CronExecution, ExecutionStatus } from '../../types/cronexecution';
import type { CronExecutionLog } from '../../types/cronexecutionlog';
import { getExecutionDetail } from '../../api/cronexecution';
import { getExecutionLogByExecutionId } from '../../api/cronexecutionlog';
import dayjs from 'dayjs';

const { Title, Paragraph } = Typography;

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
 * 日志级别颜色映射
 */
const logLevelColorMap: Record<string, string> = {
  info: 'blue',
  warn: 'orange',
  error: 'red',
  debug: 'purple',
};

/**
 * 执行记录详情页面
 */
const ExecutionDetail: React.FC = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const [execution, setExecution] = useState<CronExecution | null>(null);
  const [logs, setLogs] = useState<CronExecutionLog[]>([]);
  const [logsLoading, setLogsLoading] = useState(false);

  /**
   * 滚动到指定日志
   */
  const scrollToLog = (logId: number) => {
    const element = document.getElementById(`log-${logId}`);
    if (element) {
      element.scrollIntoView({ behavior: 'smooth', block: 'center' });
    }
  };

  /**
   * 加载执行记录详情
   */
  const loadExecutionDetail = async () => {
    if (!id) return;
    try {
      const res = await getExecutionDetail(parseInt(id, 10));
      setExecution(res.data);
    } catch (error) {
      console.error('加载执行记录详情失败:', error);
    }
  };

  /**
   * 加载执行日志
   */
  const loadLogs = async () => {
    if (!id) return;
    setLogsLoading(true);
    try {
      const res = await getExecutionLogByExecutionId(parseInt(id, 10), { page: 1, pageSize: 100 });
      setLogs(res.data.items);
    } catch (error) {
      console.error('加载执行日志失败:', error);
    } finally {
      setLogsLoading(false);
    }
  };

  useEffect(() => {
    loadExecutionDetail();
    loadLogs();
  }, [id]);

  if (!execution) {
    return (
      <Card style={{ margin: 16 }}>
        <Paragraph>加载中...</Paragraph>
      </Card>
    );
  }

  return (
    <Card
      className="cyber-card"
      style={{ margin: 16 }}
      title={
        <Space>
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate('/task/execution')}
          >
            返回
          </Button>
          <Title level={4} style={{ margin: 0 }}>执行记录详情</Title>
        </Space>
      }
    >
      <Descriptions bordered column={2}>
        <Descriptions.Item label="ID">{execution.id}</Descriptions.Item>
        <Descriptions.Item label="任务 ID">
          <Button
            type="link"
            onClick={() => navigate(`/task/${execution.task_id}`)}
          >
            #{execution.task_id}
          </Button>
        </Descriptions.Item>
        <Descriptions.Item label="计划执行时间">
          {dayjs(execution.scheduled_at).format('YYYY-MM-DD HH:mm:ss')}
        </Descriptions.Item>
        <Descriptions.Item label="状态">
          <Tag color={statusColorMap[execution.status]}>
            {execution.status}
          </Tag>
        </Descriptions.Item>
        <Descriptions.Item label="开始时间">
          {execution.started_at
            ? dayjs(execution.started_at).format('YYYY-MM-DD HH:mm:ss')
            : '-'}
        </Descriptions.Item>
        <Descriptions.Item label="完成时间">
          {execution.completed_at
            ? dayjs(execution.completed_at).format('YYYY-MM-DD HH:mm:ss')
            : '-'}
        </Descriptions.Item>
        <Descriptions.Item label="重试次数">{execution.retry_count}</Descriptions.Item>
        <Descriptions.Item label="错误信息" span={2}>
          <Paragraph copyable={!!execution.error} style={{ margin: 0 }}>
            {execution.error || '-'}
          </Paragraph>
        </Descriptions.Item>
        <Descriptions.Item label="创建时间">
          {dayjs(execution.created_at).format('YYYY-MM-DD HH:mm:ss')}
        </Descriptions.Item>
        <Descriptions.Item label="更新时间">
          {dayjs(execution.updated_at).format('YYYY-MM-DD HH:mm:ss')}
        </Descriptions.Item>
      </Descriptions>

      <Divider>执行日志</Divider>

      <div className="chat-wrapper">
        <div className="chat-messages">
          {logsLoading ? (
            <Paragraph>加载中...</Paragraph>
          ) : logs.length === 0 ? (
            <Paragraph style={{ textAlign: 'center', color: '#888' }}>暂无日志</Paragraph>
          ) : (
            <>
              {logs.map((log, index) => {
                const isEven = index % 2 === 0;
                return (
                  <div
                    key={log.id}
                    id={`log-${log.id}`}
                    className={`chat-message ${isEven ? 'chat-message-even' : 'chat-message-odd'}`}
                  >
                    <div className="chat-message-header">
                      <div className="chat-message-header-left">
                        <Tag color={logLevelColorMap[log.level.toLowerCase()] || 'default'}>
                          {log.level.toUpperCase()}
                        </Tag>
                        <span className="chat-message-from">{log.from}</span>
                      </div>
                      <Button
                        type="text"
                        size="small"
                        icon={<ClockCircleOutlined />}
                        onClick={() => scrollToLog(log.id)}
                        className="chat-time-btn"
                      >
                        {dayjs(log.created_at).format('YYYY-MM-DD HH:mm:ss')}
                      </Button>
                    </div>
                    <div className="chat-message-content">
                      <ReactMarkdown remarkPlugins={[remarkGfm]} rehypePlugins={[rehypeRaw]}>{log.message}</ReactMarkdown>
                    </div>
                  </div>
                );
              })}
            </>
          )}
        </div>
        <div className="chat-timeline">
          {logs.map((log, index) => (
            <div
              key={log.id}
              className="chat-timeline-item"
              onClick={() => scrollToLog(log.id)}
              title={dayjs(log.created_at).format('YYYY-MM-DD HH:mm:ss')}
            >
              <div className="chat-timeline-dot" />
              {index < logs.length - 1 && <div className="chat-timeline-line" />}
            </div>
          ))}
        </div>
      </div>
    </Card>
  );
};

export default ExecutionDetail;
