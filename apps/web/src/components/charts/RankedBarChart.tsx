import { Chart as ChartJS } from 'chart.js'
import { Bar } from 'react-chartjs-2'
import ChartDataLabels from 'chartjs-plugin-datalabels'
import './chartSetup'
import { chartColors, tooltipTheme } from './chartSetup'

ChartJS.register(ChartDataLabels)

interface RankedBarChartProps {
  items: { label: string; value: number }[]
  emptyMessage?: string
}

/** Horizontal ranked bar — one hue (magnitude, not identity), per
 * docs dataviz guidance: a ranked top-N list is a magnitude comparison, not
 * a case for a multi-hue categorical palette. */
export function RankedBarChart({ items, emptyMessage = 'No data yet' }: RankedBarChartProps) {
  if (items.length === 0) {
    return <p style={{ color: chartColors.ink400, fontSize: '0.85rem' }}>{emptyMessage}</p>
  }

  const height = Math.max(items.length * 34, 80)

  return (
    <div style={{ height }}>
      <Bar
        data={{
          labels: items.map((i) => i.label),
          datasets: [
            {
              data: items.map((i) => i.value),
              backgroundColor: chartColors.coral,
              borderRadius: 4,
              borderSkipped: 'left',
              maxBarThickness: 20,
            },
          ],
        }}
        options={{
          indexAxis: 'y',
          responsive: true,
          maintainAspectRatio: false,
          layout: { padding: { right: 28 } },
          scales: {
            x: {
              display: false,
              grid: { display: false },
            },
            y: {
              grid: { display: false },
              ticks: { color: chartColors.ink100, font: { family: 'Inter', size: 12 } },
            },
          },
          plugins: {
            legend: { display: false },
            tooltip: {
              ...tooltipTheme,
              callbacks: {
                label: (ctx) => `${ctx.formattedValue} plays`,
              },
            },
            datalabels: {
              anchor: 'end',
              align: 'end',
              color: chartColors.ink400,
              font: { family: 'JetBrains Mono', size: 11 },
            },
          },
        }}
      />
    </div>
  )
}
