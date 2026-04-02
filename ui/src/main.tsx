import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App.tsx'
import './globals.css' // Assuming globals.css is moved here
import { HelmetProvider } from 'react-helmet-async';
import { BrowserRouter } from 'react-router-dom';
import { TooltipProvider } from "@/components/ui/tooltip";
import { ToastProvider } from "@/components/toast";

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <HelmetProvider>
      <BrowserRouter>
        <TooltipProvider>
          <ToastProvider>
            <App />
          </ToastProvider>
        </TooltipProvider>
      </BrowserRouter>
    </HelmetProvider>
  </React.StrictMode>,
) 