import React from 'react';
import { Outlet } from 'react-router-dom';
import { Box, Container } from '@mui/material';
import Navbar from '../components/Navbar';
import Footer from '../components/Footer';

const MainLayout = () => {
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', minHeight: '100vh' }}>
      <Navbar />
      <Container component="main" sx={{ flex: 1, py: 4 }}>
        <Outlet />
      </Container>
      <Footer />
    </Box>
  );
};

export default MainLayout;
