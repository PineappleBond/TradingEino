import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  Card,
  Typography,
  Button,
  Space,
  Table,
  Select,
  DatePicker,
  Divider,
  Tag,
} from 'antd';
import { ArrowLeftOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import type { LogEntry, SystemLogLevel } from '../../types/systemlog';
import { getLogContent } from '../../api/systemlog';
import dayjs, { type Dayjs } from 'dayjs';

const { Title } = Typography;
const { RangePicker } = DatePicker;

/**
 * 分页大小选项
 */
const pageSizeOptions = ['10', '20', '50', '100'];

/**
 * 日志级别颜色映射
 */
const logLevelColorMap: Record<string, string> = {
  INFO: 'blue',
  WARN: 'orange',
  ERROR: 'red',
  DEBUG: 'purple',
};

/**
 * 系统日志文件内容页面
 */
const SystemLogDetail: React.FC = () => {
  const { filename } = useParams<{ filename: string }>();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [entries, setEntries] = useState<LogEntry[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  // 筛选条件
  const [filterLevel, setFilterLevel] = useState<SystemLogLevel | undefined>();
  const [filterTimeRange, setFilterTimeRange] = useState<[Dayjs, Dayjs] | null>(null);

  /**
   * 加载日志内容
   */
  const loadLogContent = async () => {
    if (!filename) return;
    setLoading(true);
    try {
      const params: any = {
        filename,
        page,
        pageSize,
      };

      if (filterLevel) {
        params.level = filterLevel;
      }

      if (filterTimeRange && filterTimeRange[0] && filterTimeRange[1]) {
        params.start_time = filterTimeRange[0].format('YYYY-MM-DD HH:mm:ss');
        params.end_time = filterTimeRange[1].format('YYYY-MM-DD HH:mm:ss');
      }

      const res = await getLogContent(params);
      setEntries(res.data.items);
      setTotal(res.data.page.total);
    } catch (error) {
      console.error('加载日志内容失败:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadLogContent();
  }, [filename, page, pageSize, filterLevel]);

  /**
   * 表格列配置
   */
  const columns: ColumnsType<LogEntry> = [
    {
      title: '级别',
      dataIndex: 'level',
      width: 80,
      render: (level: SystemLogLevel) => (
        <Tag color={logLevelColorMap[level] || 'default'}>
          {level}
        </Tag>
      ),
    },
    {
      title: '时间',
      dataIndex: 'time',
      width: 180,
      render: (time: string) => dayjs(time).format('YYYY-MM-DD HH:mm:ss.SSS'),
    },
    {
      title: '消息',
      dataIndex: 'msg',
      ellipsis: true,
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
          { label: 'INFO', value: 'INFO' },
          { label: 'WARN', value: 'WARN' },
          { label: 'ERROR', value: 'ERROR' },
          { label: 'DEBUG', value: 'DEBUG' },
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
          setFilterLevel(undefined);
          setFilterTimeRange(null);
          setPage(1);
        }}
      >
        重置筛选
      </Button>
    </Space>
  );

  return (
    <Card
      className="cyber-card"
      style={{ margin: 16 }}
      title={
        <Space>
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate('/log/system')}
          >
            返回
          </Button>
          <Title level={4} style={{ margin: 0 }}>
            日志文件：{filename}
          </Title>
        </Space>
      }
    >
      {filterSection}

      <Divider style={{ margin: '12px 0' }} />

      <Table
        rowKey={(record) => `${record.time}-${record.msg}`}
        columns={columns}
        dataSource={entries}
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
        scroll={{ x: 800 }}
      />
    </Card>
  );
};

export default SystemLogDetail;
