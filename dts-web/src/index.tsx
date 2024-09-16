import React from 'react';
import ReactDOM from 'react-dom/client';
import './index.css';
import App from './App';
import { AuthKitProvider } from '@workos-inc/authkit-react';

const root = ReactDOM.createRoot(
  document.getElementById('root') as HTMLElement
);
root.render(
  <React.StrictMode>
    <AuthKitProvider clientId="client_01J7E5R6JNRYBG12BJQ8J7DGK3" apiHostname="rousing-editor-46-staging.authkit.app">
      <App />
    </AuthKitProvider>
  </React.StrictMode>
);

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
