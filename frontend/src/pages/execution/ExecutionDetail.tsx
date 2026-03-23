import { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
  Card,
  Button,
  Space,
  Tag,
  Typography,
  Drawer,
  message,
  Descriptions,
} from 'antd';
import { ArrowLeftOutlined, ClockCircleOutlined, CopyOutlined, PicCenterOutlined } from '@ant-design/icons';
import html2canvas from 'html2canvas';
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
  const [drawerOpen, setDrawerOpen] = useState(false);

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
   * 复制日志原文
   */
  const handleCopyText = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      message.success('已复制原文');
    } catch (error) {
      message.error('复制失败');
    }
  };

  /**
   * 复制日志为图片
   */
  const handleCopyImage = async (logId: number) => {
    try {
      const element = document.getElementById(`log-${logId}`);
      if (!element) return;

      const canvas = await html2canvas(element, {
        backgroundColor: '#252535',
        scale: 2,
        logging: false,
        useCORS: true,
      });

      canvas.toBlob((blob) => {
        if (blob) {
          const item = new ClipboardItem({ 'image/png': blob });
          navigator.clipboard.write([item]).then(() => {
            message.success('已复制为图片');
          }).catch(() => {
            // 如果剪贴板 API 失败，则下载图片
            const link = document.createElement('a');
            link.href = canvas.toDataURL('image/png');
            link.download = `log-${logId}.png`;
            link.click();
            message.success('已下载图片');
          });
        }
      });
    } catch (error) {
      message.error('截图失败');
    }
  };

  /**
   * 根据发送者生成头像颜色
   */
  const getAvatarColor = (from: string): string => {
    const colors = ['#bf00ff', '#00ffff', '#ff0080', '#00ff80', '#ff8000', '#8080ff'];
    const hash = from.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0);
    return colors[hash % colors.length];
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
      style={{ margin: 24, height: 'calc(100% - 48px)' }}
      styles={{ body: { padding: 0, height: '100%' } }}
      title={
        <Space>
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate('/task/execution')}
          >
            返回
          </Button>
          <Title level={4} style={{ margin: 0 }}>执行记录详情</Title>
          <Button
            icon={<CopyOutlined />}
            onClick={() => setDrawerOpen(true)}
          >
            详情
          </Button>
        </Space>
      }
    >
      <div className="chat-wrapper">
        <div className="chat-messages">
          {logsLoading ? (
            <Paragraph>加载中...</Paragraph>
          ) : logs.length === 0 ? (
            <Paragraph style={{ textAlign: 'center', color: '#888' }}>暂无日志</Paragraph>
          ) : (
            <>
              {logs.map((log) => (
                <div
                  key={log.id}
                  id={`log-${log.id}`}
                  className="chat-message-wrapper"
                >
                  <div
                    className="chat-avatar"
                    style={{ background: getAvatarColor(log.from) }}
                  >
                    {log.from.charAt(0).toUpperCase()}
                  </div>
                  <div className="chat-message">
                    <div className="chat-message-actions">
                      <button className="chat-action-btn" onClick={() => handleCopyText(log.message)}>
                        <CopyOutlined /> 复制原文
                      </button>
                      <button className="chat-action-btn" onClick={() => handleCopyImage(log.id)}>
                        <PicCenterOutlined /> 截图
                      </button>
                    </div>
                    <div className="chat-message-content-wrapper">
                      <div className="chat-message-header">
                        <div className="chat-message-header-left">
                          <Tag color={logLevelColorMap[log.level.toLowerCase()] || 'default'}>
                            {log.level.toUpperCase()}
                          </Tag>
                          <span className="chat-message-from">{log.from}</span>
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
                      </div>
                      <div className="chat-message-content">
                        <ReactMarkdown remarkPlugins={[remarkGfm]} rehypePlugins={[rehypeRaw]}>{log.message}</ReactMarkdown>
                      </div>
                    </div>
                  </div>
                </div>
              ))}
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
              <div
                className="chat-timeline-avatar"
                style={{ background: getAvatarColor(log.from) }}
              >
                {log.from.charAt(0).toUpperCase()}
              </div>
              {index < logs.length - 1 && <div className="chat-timeline-line" />}
            </div>
          ))}
        </div>
      </div>

      <Drawer
        title="执行详情"
        placement="right"
        width={480}
        open={drawerOpen}
        onClose={() => setDrawerOpen(false)}
      >
        <Descriptions bordered column={1}>
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
          <Descriptions.Item label="错误信息">
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
      </Drawer>
    </Card>
  );
};

export default ExecutionDetail;
