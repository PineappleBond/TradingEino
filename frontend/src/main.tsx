import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { ConfigProvider } from 'antd'
import { cyberpunkTheme } from './styles/theme'
import './styles/global.css'
import App from './App.tsx'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <ConfigProvider theme={cyberpunkTheme}>
      <App />
    </ConfigProvider>
  </StrictMode>,
)
