import { useState, useEffect } from 'react';
import { useNavigate, useParams, useSearchParams } from 'react-router-dom';
import {
  Card,
  Form,
  Input,
  Select,
  DatePicker,
  Button,
  Space,
  message,
  Typography,
  Divider,
} from 'antd';
import { ArrowLeftOutlined, SaveOutlined } from '@ant-design/icons';
import type { Dayjs } from 'dayjs';
import dayjs from 'dayjs';
import type { CreateTaskRequest, UpdateTaskRequest } from '../../types/crontask';
import { createTask, updateTask, getTaskDetail } from '../../api/crontask';

const { Title } = Typography;
const { RangePicker } = DatePicker;

/**
 * 任务表单组件（创建/编辑复用）
 */
const TaskForm: React.FC = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const [searchParams] = useSearchParams();
  const copyFrom = searchParams.get('copyFrom');
  const isEdit = !!id;
  const isCopy = !!copyFrom;
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [validRange, setValidRange] = useState<[Dayjs, Dayjs] | null>(null);
  const [taskType, setTaskType] = useState<'once' | 'recurring'>('once');
  const [nextExecutionTime, setNextExecutionTime] = useState<Dayjs | null>(null);

  /**
   * 加载任务详情（编辑模式或拷贝模式）
   */
  useEffect(() => {
    if (isEdit && id) {
      loadTaskDetail();
    } else if (isCopy && copyFrom) {
      loadTaskForCopy(copyFrom);
    }
  }, [id, isEdit, copyFrom, isCopy]);

  const loadTaskDetail = async () => {
    try {
      const res = await getTaskDetail(parseInt(id!, 10));
      const task = res.data;

      // 设置任务类型（在填充表单之前，避免渲染问题）
      const type = task.type as 'once' | 'recurring';
      setTaskType(type);

      // 设置下次执行时间（仅一次性任务）
      if (type === 'once' && task.next_execution_at) {
        setNextExecutionTime(dayjs(task.next_execution_at));
      }

      // 填充表单值
      form.setFieldsValue({
        name: task.name + ' - 副本',
        spec: task.spec,
        type: type,
        exec_type: task.exec_type,
        symbol: JSON.parse(task.raw)?.symbol || '',
        valid_from: dayjs(task.valid_from),
        valid_until: dayjs(task.valid_until),
        enabled: task.enabled,
        max_retries: task.max_retries,
        timeout_seconds: task.timeout_seconds,
      });

      // 设置日期范围
      setValidRange([dayjs(task.valid_from), dayjs(task.valid_until)]);
    } catch (error) {
      console.error('加载任务详情失败:', error);
      message.error('加载任务详情失败');
    }
  };

  /**
   * 加载任务数据用于拷贝
   */
  const loadTaskForCopy = async (copyFromId: string) => {
    try {
      const res = await getTaskDetail(parseInt(copyFromId, 10));
      const task = res.data;

      // 设置任务类型（在填充表单之前，避免渲染问题）
      const type = task.type as 'once' | 'recurring';
      setTaskType(type);

      // 设置下次执行时间（仅一次性任务）
      if (type === 'once' && task.next_execution_at) {
        setNextExecutionTime(dayjs(task.next_execution_at));
      }

      // 填充表单值（复制所有参数，名称添加副本后缀）
      form.setFieldsValue({
        name: task.name + ' - 副本',
        spec: task.spec,
        type: type,
        exec_type: task.exec_type,
        symbol: JSON.parse(task.raw)?.symbol || '',
        valid_from: dayjs(task.valid_from),
        valid_until: dayjs(task.valid_until),
        enabled: task.enabled,
        max_retries: task.max_retries,
        timeout_seconds: task.timeout_seconds,
      });

      // 设置日期范围
      setValidRange([dayjs(task.valid_from), dayjs(task.valid_until)]);

      message.success('已复制任务参数，请修改后保存');
    } catch (error) {
      console.error('加载任务数据失败:', error);
      message.error('加载任务数据失败');
    }
  };

  /**
   * 处理任务类型切换
   */
  const handleTypeChange = (value: 'once' | 'recurring') => {
    setTaskType(value);
    // 切换类型时清空对应字段
    if (value === 'once') {
      form.setFieldsValue({ spec: undefined });
    } else {
      setNextExecutionTime(null);
    }
  };

  /**
   * 提交表单
   */
  const handleSubmit = async (values: any) => {
    setLoading(true);
    try {
      // 构建 raw JSON 字符串
      const raw = JSON.stringify({ symbol: values.symbol });

      // 转换日期格式为 RFC3339
      const validFrom = values.valid_from.format('YYYY-MM-DDTHH:mm:ssZ');
      const validUntil = values.valid_until.format('YYYY-MM-DDTHH:mm:ssZ');

      const data: CreateTaskRequest | UpdateTaskRequest = {
        name: values.name,
        type: values.type,
        exec_type: values.exec_type,
        raw,
        valid_from: validFrom,
        valid_until: validUntil,
        enabled: values.enabled,
        max_retries: values.max_retries,
        timeout_seconds: values.timeout_seconds,
      };

      // 根据任务类型设置 spec 或 next_execution_at
      if (values.type === 'once') {
        // 一次性任务：传入 next_execution_at，不传 spec
        if (nextExecutionTime) {
          (data as any).next_execution_at = nextExecutionTime.format('YYYY-MM-DDTHH:mm:ssZ');
        }
      } else {
        // 周期性任务：传入 spec
        (data as any).spec = values.spec;
      }

      if (isEdit) {
        await updateTask(parseInt(id!, 10), data);
        message.success('更新成功');
      } else {
        await createTask(data as CreateTaskRequest);
        message.success('创建成功');
      }

      // 关闭弹窗，返回列表
      navigate('/task');
    } catch (error) {
      console.error('提交失败:', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card
      className="cyber-card"
      style={{ margin: 16 }}
      title={
        <Space>
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/task')}>
            返回
          </Button>
          <Title level={4} style={{ margin: 0 }}>
            {isEdit ? '编辑任务' : isCopy ? '拷贝任务' : '创建任务'}
          </Title>
        </Space>
      }
    >
      <Form
        form={form}
        layout="vertical"
        onFinish={handleSubmit}
        style={{ maxWidth: 800 }}
        initialValues={{
          exec_type: 'OKXWatcher',
          enabled: true,
          max_retries: 3,
          timeout_seconds: 300,
        }}
      >
        <Divider orientation="left">基本信息</Divider>

        <Form.Item
          name="name"
          label="任务名称"
          rules={[
            { required: true, message: '请输入任务名称' },
            { max: 100, message: '任务名称不能超过 100 个字符' },
          ]}
        >
          <Input placeholder="请输入任务名称" />
        </Form.Item>

        <Form.Item
          name="type"
          label="任务类型"
          rules={[{ required: true, message: '请选择任务类型' }]}
        >
          <Select
            placeholder="请选择任务类型"
            onChange={handleTypeChange}
          >
            <Select.Option value="once">一次性</Select.Option>
            <Select.Option value="recurring">周期性</Select.Option>
          </Select>
        </Form.Item>

        {/* 周期性任务：显示 Cron 表达式 */}
        {taskType === 'recurring' && (
          <Form.Item
            name="spec"
            label="Cron 表达式"
            rules={[
              { required: true, message: '请输入 Cron 表达式' },
              { max: 50, message: 'Cron 表达式不能超过 50 个字符' },
            ]}
            extra="例如：*/5 * * * * 表示每 5 分钟执行一次"
          >
            <Input placeholder="请输入 Cron 表达式" />
          </Form.Item>
        )}

        {/* 一次性任务：显示下次执行时间 */}
        {taskType === 'once' && (
          <Form.Item
            label="下次执行时间"
            required
            rules={[{ required: true, message: '请选择下次执行时间' }]}
          >
            <DatePicker
              value={nextExecutionTime}
              onChange={(date) => setNextExecutionTime(date)}
              showTime
              format="YYYY-MM-DD HH:mm:ss"
              style={{ width: '100%' }}
              placeholder="请选择下次执行时间"
            />
          </Form.Item>
        )}

        <Form.Item
          name="exec_type"
          label="执行类型"
          rules={[{ required: true, message: '请选择执行类型' }]}
        >
          <Select placeholder="请选择执行类型" disabled>
            <Select.Option value="OKXWatcher">OKXWatcher</Select.Option>
          </Select>
        </Form.Item>

        <Divider orientation="left">OKX 配置</Divider>

        <Form.Item
          name="symbol"
          label="交易对 (Symbol)"
          rules={[{ required: true, message: '请输入交易对' }]}
          extra="例如：ETH-USDT-SWAP"
        >
          <Input placeholder="请输入交易对，如 ETH-USDT-SWAP" />
        </Form.Item>

        <Divider orientation="left">有效期设置</Divider>

        <Form.Item
          label="有效期"
          required
          rules={[{ required: true, message: '请选择有效期' }]}
        >
          <RangePicker
            value={validRange}
            onChange={(dates) => {
              if (dates && dates[0] && dates[1]) {
                setValidRange([dates[0], dates[1]]);
                form.setFieldsValue({
                  valid_from: dates[0],
                  valid_until: dates[1],
                });
              }
            }}
            showTime={{ format: 'HH:mm:ss' }}
            format="YYYY-MM-DD HH:mm:ss"
            style={{ width: '100%' }}
          />
        </Form.Item>

        {/* 隐藏的日期字段，用于提交 */}
        <Form.Item name="valid_from" hidden>
          <Input />
        </Form.Item>
        <Form.Item name="valid_until" hidden>
          <Input />
        </Form.Item>

        <Divider orientation="left">高级配置</Divider>

        <Form.Item
          name="enabled"
          label="是否启用"
          valuePropName="checked"
        >
          <Select>
            <Select.Option value={true}>启用</Select.Option>
            <Select.Option value={false}>禁用</Select.Option>
          </Select>
        </Form.Item>

        <Form.Item
          name="max_retries"
          label="最大重试次数"
          rules={[{ type: 'number', min: 0 }]}
        >
          <Input type="number" placeholder="默认：3" />
        </Form.Item>

        <Form.Item
          name="timeout_seconds"
          label="超时时间 (秒)"
          rules={[{ type: 'number', min: 1 }]}
        >
          <Input type="number" placeholder="默认：300" />
        </Form.Item>

        <Form.Item>
          <Space>
            <Button
              type="primary"
              htmlType="submit"
              icon={<SaveOutlined />}
              loading={loading}
            >
              {isEdit ? '保存' : '创建'}
            </Button>
            <Button onClick={() => navigate('/task')}>
              取消
            </Button>
          </Space>
        </Form.Item>
      </Form>
    </Card>
  );
};

export default TaskForm;
