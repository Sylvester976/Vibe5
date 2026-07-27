import {
  BarController,
  BarElement,
  CategoryScale,
  Chart as ChartJS,
  Legend,
  LinearScale,
  Tooltip,
} from 'chart.js'

ChartJS.register(BarController, BarElement, CategoryScale, LinearScale, Tooltip, Legend)

// Shared token values (kept in sync with src/styles/tokens.css — Chart.js
// can't read CSS custom properties directly in canvas fill styles).
export const chartColors = {
  coral: '#ff6b4a',
  ink100: '#f2efec',
  ink400: '#8b8790',
  bgRaised: '#1d1a21',
  gridline: 'rgba(255, 255, 255, 0.08)',
}

export const tooltipTheme = {
  backgroundColor: chartColors.bgRaised,
  titleColor: chartColors.ink400,
  bodyColor: chartColors.ink100,
  borderColor: chartColors.gridline,
  borderWidth: 1,
  padding: 10,
  displayColors: false,
  titleFont: { family: 'Inter', size: 11 },
  bodyFont: { family: 'JetBrains Mono', size: 13, weight: 'bold' as const },
}
