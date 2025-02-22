import React from 'react';
import { Box, Container, Typography, Link } from '@mui/material';

const Footer = () => {
  return (
    <Box
      component="footer"
      sx={{
        py: 3,
        px: 2,
        mt: 'auto',
        backgroundColor: (theme) =>
          theme.palette.mode === 'light'
            ? theme.palette.grey[200]
            : theme.palette.grey[800],
      }}
    >
      <Container maxWidth="sm">
        <Typography variant="body1" align="center">
          © {new Date().getFullYear()} Blog Platform. All rights reserved.
        </Typography>
        <Typography variant="body2" color="text.secondary" align="center">
          {'Built with '}
          <Link color="inherit" href="https://golang.org/">
            Go
          </Link>
          {' and '}
          <Link color="inherit" href="https://reactjs.org/">
            React
          </Link>
        </Typography>
      </Container>
    </Box>
  );
};

export default Footer;
