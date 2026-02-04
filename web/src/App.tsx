import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import Home from '@/pages/Home'

import Login from '@/pages/auth/Login'
import Register from '@/pages/auth/Register'

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />

        {/* Fallback */}
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  )
}

export default App
