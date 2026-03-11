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
      backgroundColor: 'rgba(18, 18, 24, 0.9)',
      borderColor: 'rgba(255, 255, 255, 0.1)',
      textStyle: {
        color: 'rgba(255, 255, 255, 0.88)',
      },
    },
    legend: {
      data: ['成功', '失败'],
      bottom: 0,
      textStyle: {
        color: 'rgba(255, 255, 255, 0.65)',
      },
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
          color: 'rgba(255, 255, 255, 0.1)',
        },
      },
      axisLabel: {
        color: 'rgba(255, 255, 255, 0.45)',
      },
      axisTick: {
        lineStyle: {
          color: 'rgba(255, 255, 255, 0.1)',
        },
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
        color: 'rgba(255, 255, 255, 0.45)',
      },
      splitLine: {
        lineStyle: {
          color: 'rgba(255, 255, 255, 0.04)',
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
          color: '#73d13d',
        },
        itemStyle: {
          color: '#73d13d',
          borderColor: '#1a1a24',
          borderWidth: 2,
        },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(115, 209, 61, 0.25)' },
              { offset: 1, color: 'rgba(115, 209, 61, 0.02)' },
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
          color: '#ff7875',
        },
        itemStyle: {
          color: '#ff7875',
          borderColor: '#1a1a24',
          borderWidth: 2,
        },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(255, 120, 117, 0.25)' },
              { offset: 1, color: 'rgba(255, 120, 117, 0.02)' },
            ],
          },
        },
        data: data?.map((item) => item.failed) || [],
      },
    ],
  }), [data])

  return (
    <Card
      title="任务执行趋势 (24小时)"
      loading={loading}
      style={{
        background: 'rgba(26, 26, 36, 0.6)',
        border: '1px solid rgba(255, 255, 255, 0.06)',
        borderRadius: 12,
      }}
    >
      <ReactECharts
        option={option}
        style={{ height: 300 }}
        opts={{ renderer: 'svg' }}
      />
    </Card>
  )
}
