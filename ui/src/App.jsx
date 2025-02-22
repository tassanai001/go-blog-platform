import React from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { ThemeProvider, CssBaseline } from '@mui/material';
import { Provider } from 'react-redux';
import theme from './theme';
import store from './store';

// Layouts
import MainLayout from './layouts/MainLayout';

// Pages
import Home from './pages/Home';
import Login from './pages/Login';
import Register from './pages/Register';
import CreatePost from './pages/CreatePost';
import EditPost from './pages/EditPost';
import ViewPost from './pages/ViewPost';
import Profile from './pages/Profile';

function App() {
  return (
    <Provider store={store}>
      <ThemeProvider theme={theme}>
        <CssBaseline />
        <BrowserRouter>
          <Routes>
            <Route path="/" element={<MainLayout />}>
              <Route index element={<Home />} />
              <Route path="login" element={<Login />} />
              <Route path="register" element={<Register />} />
              <Route path="posts/create" element={<CreatePost />} />
              <Route path="posts/:id" element={<ViewPost />} />
              <Route path="posts/:id/edit" element={<EditPost />} />
              <Route path="profile" element={<Profile />} />
            </Route>
          </Routes>
        </BrowserRouter>
      </ThemeProvider>
    </Provider>
  );
}

export default App;
