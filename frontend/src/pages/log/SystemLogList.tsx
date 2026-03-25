import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Table,
  Button,
  Card,
  Typography,
  Space,
} from 'antd';
import { FileTextOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import type { LogFileInfo } from '../../types/systemlog';
import { getLogFileList } from '../../api/systemlog';
import dayjs from 'dayjs';

const { Title } = Typography;

/**
 * 分页大小选项
 */
const pageSizeOptions = ['10', '20', '50', '100'];

/**
 * 系统日志文件列表页面
 */
const SystemLogList: React.FC = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [files, setFiles] = useState<LogFileInfo[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  /**
   * 加载日志文件列表
   */
  const loadLogFiles = async () => {
    setLoading(true);
    try {
      const res = await getLogFileList({ page, pageSize });
      setFiles(res.data.items);
      setTotal(res.data.page.total);
    } catch (error) {
      console.error('加载日志文件列表失败:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadLogFiles();
  }, [page, pageSize]);

  /**
   * 表格列配置
   */
  const columns: ColumnsType<LogFileInfo> = [
    {
      title: '文件名',
      dataIndex: 'filename',
      width: 200,
      render: (filename: string) => (
        <Space>
          <FileTextOutlined />
          <span style={{ fontFamily: 'monospace' }}>{filename}</span>
        </Space>
      ),
    },
    {
      title: '文件大小',
      dataIndex: 'size',
      width: 100,
      render: (size: number) => {
        if (size < 1024) return `${size} B`;
        if (size < 1024 * 1024) return `${(size / 1024).toFixed(2)} KB`;
        if (size < 1024 * 1024 * 1024) return `${(size / 1024 / 1024).toFixed(2)} MB`;
        return `${(size / 1024 / 1024 / 1024).toFixed(2)} GB`;
      },
    },
    {
      title: '行数',
      dataIndex: 'line_count',
      width: 80,
      align: 'right',
    },
    {
      title: '首条日志时间',
      dataIndex: 'first_log_time',
      width: 160,
      render: (time?: string) =>
        time ? dayjs(time).format('YYYY-MM-DD HH:mm:ss') : '-',
    },
    {
      title: '末条日志时间',
      dataIndex: 'last_log_time',
      width: 160,
      render: (time?: string) =>
        time ? dayjs(time).format('YYYY-MM-DD HH:mm:ss') : '-',
    },
    {
      title: '修改时间',
      dataIndex: 'mod_time',
      width: 160,
      render: (time: string) => dayjs(time).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: '操作',
      key: 'action',
      width: 100,
      fixed: 'right',
      render: (_: unknown, record: LogFileInfo) => (
        <Button
          type="link"
          icon={<FileTextOutlined />}
          onClick={() => navigate(`/log/system/${record.filename}`)}
        >
          查看内容
        </Button>
      ),
    },
  ];

  return (
    <Card className="cyber-card" style={{ margin: 16 }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0, color: '#bf00ff' }}>系统日志文件列表</Title>
      </div>

      <Table
        rowKey="filename"
        columns={columns}
        dataSource={files}
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
        scroll={{ x: 1200 }}
      />
    </Card>
  );
};

export default SystemLogList;
