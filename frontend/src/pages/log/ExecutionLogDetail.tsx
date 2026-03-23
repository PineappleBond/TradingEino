import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  Card,
  Descriptions,
  Button,
  Space,
  Tag,
  Typography,
} from 'antd';
import { ArrowLeftOutlined } from '@ant-design/icons';
import type { CronExecutionLog } from '../../types/cronexecutionlog';
import { getExecutionLogDetail } from '../../api/cronexecutionlog';
import dayjs from 'dayjs';

const { Title, Paragraph } = Typography;

/**
 * 执行日志详情页面
 */
const ExecutionLogDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [log, setLog] = useState<CronExecutionLog | null>(null);

  /**
   * 加载日志详情
   */
  const loadLogDetail = async () => {
    if (!id) return;
    try {
      const res = await getExecutionLogDetail(parseInt(id, 10));
      setLog(res.data);
    } catch (error) {
      console.error('加载日志详情失败:', error);
    }
  };

  useEffect(() => {
    loadLogDetail();
  }, [id]);

  if (!log) {
    return (
      <Card style={{ margin: 16 }}>
        <Paragraph>加载中...</Paragraph>
      </Card>
    );
  }

  /**
   * 日志级别颜色映射
   */
  const logLevelColorMap: Record<string, string> = {
    info: 'blue',
    warn: 'orange',
    error: 'red',
    debug: 'purple',
  };

  return (
    <Card
      className="cyber-card"
      style={{ margin: 16 }}
      title={
        <Space>
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate('/log/execution')}
          >
            返回
          </Button>
          <Title level={4} style={{ margin: 0 }}>执行日志详情</Title>
        </Space>
      }
    >
      <Descriptions bordered column={1}>
        <Descriptions.Item label="ID">{log.id}</Descriptions.Item>
        <Descriptions.Item label="执行 ID">
          <Button
            type="link"
            onClick={() => navigate(`/task/execution/${log.execution_id}`)}
          >
            #{log.execution_id}
          </Button>
        </Descriptions.Item>
        <Descriptions.Item label="级别">
          <Tag color={logLevelColorMap[log.level.toLowerCase()] || 'default'}>
            {log.level.toUpperCase()}
          </Tag>
        </Descriptions.Item>
        <Descriptions.Item label="来源">{log.from}</Descriptions.Item>
        <Descriptions.Item label="消息">
          <Paragraph style={{ margin: 0 }}>{log.message}</Paragraph>
        </Descriptions.Item>
        <Descriptions.Item label="创建时间">
          {dayjs(log.created_at).format('YYYY-MM-DD HH:mm:ss')}
        </Descriptions.Item>
        <Descriptions.Item label="更新时间">
          {dayjs(log.updated_at).format('YYYY-MM-DD HH:mm:ss')}
        </Descriptions.Item>
      </Descriptions>
    </Card>
  );
};

export default ExecutionLogDetail;
