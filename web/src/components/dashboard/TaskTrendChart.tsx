import { useMemo } from 'react'
import { Card } from 'antd'
import ReactECharts from 'echarts-for-react'
import type { TaskTrendItem } from '@/types'

interface TaskTrendChartProps {
  data: TaskTrendItem[] | undefined
  loading?: boolean
}

export default function TaskTrendChart({ data, loading }: TaskTrendChartProps) {
  const option = useMemo(() => ({
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'cross',
      },
    },
    legend: {
      data: ['成功', '失败'],
      bottom: 0,
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '15%',
      top: '10%',
      containLabel: true,
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: data?.map((item) => item.time) || [],
      axisLine: {
        lineStyle: {
          color: '#d9d9d9',
        },
      },
      axisLabel: {
        color: '#8c8c8c',
      },
    },
    yAxis: {
      type: 'value',
      minInterval: 1,
      axisLine: {
        show: false,
      },
      axisTick: {
        show: false,
      },
      axisLabel: {
        color: '#8c8c8c',
      },
      splitLine: {
        lineStyle: {
          color: '#f0f0f0',
        },
      },
    },
    series: [
      {
        name: '成功',
        type: 'line',
        smooth: true,
        symbol: 'circle',
        symbolSize: 6,
        lineStyle: {
          width: 2,
          color: '#52c41a',
        },
        itemStyle: {
          color: '#52c41a',
        },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(82, 196, 26, 0.3)' },
              { offset: 1, color: 'rgba(82, 196, 26, 0.05)' },
            ],
          },
        },
        data: data?.map((item) => item.completed) || [],
      },
      {
        name: '失败',
        type: 'line',
        smooth: true,
        symbol: 'circle',
        symbolSize: 6,
        lineStyle: {
          width: 2,
          color: '#ff4d4f',
        },
        itemStyle: {
          color: '#ff4d4f',
        },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(255, 77, 79, 0.3)' },
              { offset: 1, color: 'rgba(255, 77, 79, 0.05)' },
            ],
          },
        },
        data: data?.map((item) => item.failed) || [],
      },
    ],
  }), [data])

  return (
    <Card title="任务执行趋势 (24小时)" loading={loading}>
      <ReactECharts
        option={option}
        style={{ height: 300 }}
        opts={{ renderer: 'svg' }}
      />
    </Card>
  )
}
