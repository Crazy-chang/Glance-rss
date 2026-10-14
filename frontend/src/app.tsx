import { PropsWithChildren } from 'react'
import { SettingProvider } from './store/setting'
import './app.scss'

function App({ children }: PropsWithChildren<any>) {
  return <SettingProvider>{children}</SettingProvider>
}

export default App
