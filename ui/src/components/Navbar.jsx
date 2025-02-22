import React from 'react';
import { useNavigate, Link as RouterLink } from 'react-router-dom';
import { useSelector, useDispatch } from 'react-redux';
import {
  AppBar,
  Toolbar,
  Typography,
  Button,
  IconButton,
  Box,
  Avatar,
  Menu,
  MenuItem,
} from '@mui/material';
import { AccountCircle } from '@mui/icons-material';
import { logout } from '../store/slices/authSlice';

const Navbar = () => {
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const { isAuthenticated, user } = useSelector((state) => state.auth);
  const [anchorEl, setAnchorEl] = React.useState(null);

  const handleMenu = (event) => {
    setAnchorEl(event.currentTarget);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };

  const handleLogout = () => {
    dispatch(logout());
    handleClose();
    navigate('/');
  };

  return (
    <AppBar position="static">
      <Toolbar>
        <Typography
          variant="h6"
          component={RouterLink}
          to="/"
          sx={{ flexGrow: 1, textDecoration: 'none', color: 'inherit' }}
        >
          Blog Platform
        </Typography>

        {isAuthenticated ? (
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
            <Button
              color="inherit"
              component={RouterLink}
              to="/posts/create"
            >
              Create Post
            </Button>
            <IconButton
              size="large"
              onClick={handleMenu}
              color="inherit"
            >
              {user?.profile?.avatar ? (
                <Avatar src={user.profile.avatar.path} alt={user.username} />
              ) : (
                <AccountCircle />
              )}
            </IconButton>
            <Menu
              anchorEl={anchorEl}
              open={Boolean(anchorEl)}
              onClose={handleClose}
            >
              <MenuItem
                component={RouterLink}
                to="/profile"
                onClick={handleClose}
              >
                Profile
              </MenuItem>
              <MenuItem onClick={handleLogout}>Logout</MenuItem>
            </Menu>
          </Box>
        ) : (
          <Box sx={{ display: 'flex', gap: 1 }}>
            <Button
              color="inherit"
              component={RouterLink}
              to="/login"
            >
              Login
            </Button>
            <Button
              color="inherit"
              component={RouterLink}
              to="/register"
            >
              Register
            </Button>
          </Box>
        )}
      </Toolbar>
    </AppBar>
  );
};

export default Navbar;
