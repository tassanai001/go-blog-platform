import React, { useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import {
  Box,
  Paper,
  TextField,
  Button,
  Typography,
  Avatar,
  Grid,
  Alert,
} from '@mui/material';
import { updateProfile } from '../store/slices/authSlice';
import { uploadMedia } from '../store/slices/mediaSlice';
import ImageUpload from '../components/ImageUpload';
import config from '../config';

const Profile = () => {
  const dispatch = useDispatch();
  const { user, loading, error } = useSelector((state) => state.auth);

  const [formData, setFormData] = useState({
    username: user?.username || '',
    email: user?.email || '',
    bio: user?.profile?.bio || '',
    avatar: user?.profile?.avatar || null,
    cover_image: user?.profile?.cover_image || null,
  });

  const [updateSuccess, setUpdateSuccess] = useState(false);

  const handleChange = (e) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value,
    });
    setUpdateSuccess(false);
  };

  const handleAvatarUpload = async (file) => {
    try {
      const formData = new FormData();
      formData.append('file', file);
      const response = await dispatch(uploadMedia(formData)).unwrap();
      setFormData((prev) => ({
        ...prev,
        avatar: response,
      }));
    } catch (error) {
      console.error('Failed to upload avatar:', error);
    }
  };

  const handleCoverImageUpload = async (file) => {
    try {
      const formData = new FormData();
      formData.append('file', file);
      const response = await dispatch(uploadMedia(formData)).unwrap();
      setFormData((prev) => ({
        ...prev,
        cover_image: response,
      }));
    } catch (error) {
      console.error('Failed to upload cover image:', error);
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      await dispatch(updateProfile(formData)).unwrap();
      setUpdateSuccess(true);
    } catch (error) {
      console.error('Failed to update profile:', error);
    }
  };

  return (
    <Box maxWidth="md" sx={{ mx: 'auto', py: 4 }}>
      <Paper sx={{ p: 4 }}>
        <Typography variant="h4" gutterBottom>
          Profile Settings
        </Typography>

        {error && (
          <Alert severity="error" sx={{ mb: 3 }}>
            {error}
          </Alert>
        )}

        {updateSuccess && (
          <Alert severity="success" sx={{ mb: 3 }}>
            Profile updated successfully!
          </Alert>
        )}

        <Box component="form" onSubmit={handleSubmit}>
          <Grid container spacing={4}>
            <Grid item xs={12}>
              <Box
                sx={{
                  position: 'relative',
                  height: 200,
                  mb: 4,
                  borderRadius: 1,
                  overflow: 'hidden',
                }}
              >
                {formData.cover_image ? (
                  <Box
                    component="img"
                    src={`${config.mediaUrl}/${formData.cover_image.path}`}
                    alt="Cover"
                    sx={{
                      width: '100%',
                      height: '100%',
                      objectFit: 'cover',
                    }}
                  />
                ) : (
                  <Box
                    sx={{
                      width: '100%',
                      height: '100%',
                      bgcolor: 'grey.200',
                    }}
                  />
                )}
                <Box sx={{ position: 'absolute', bottom: 16, right: 16 }}>
                  <ImageUpload
                    onUpload={handleCoverImageUpload}
                    buttonText="Change Cover"
                  />
                </Box>
              </Box>

              <Box
                sx={{
                  display: 'flex',
                  alignItems: 'center',
                  mb: 4,
                }}
              >
                <Avatar
                  src={
                    formData.avatar
                      ? `${config.mediaUrl}/${formData.avatar.path}`
                      : undefined
                  }
                  sx={{ width: 100, height: 100, mr: 2 }}
                />
                <ImageUpload
                  onUpload={handleAvatarUpload}
                  buttonText="Change Avatar"
                />
              </Box>
            </Grid>

            <Grid item xs={12}>
              <TextField
                fullWidth
                label="Username"
                name="username"
                value={formData.username}
                onChange={handleChange}
                required
                sx={{ mb: 3 }}
              />
            </Grid>

            <Grid item xs={12}>
              <TextField
                fullWidth
                label="Email"
                name="email"
                type="email"
                value={formData.email}
                onChange={handleChange}
                required
                sx={{ mb: 3 }}
              />
            </Grid>

            <Grid item xs={12}>
              <TextField
                fullWidth
                label="Bio"
                name="bio"
                value={formData.bio}
                onChange={handleChange}
                multiline
                rows={4}
                sx={{ mb: 3 }}
              />
            </Grid>

            <Grid item xs={12}>
              <Button
                type="submit"
                variant="contained"
                color="primary"
                disabled={loading}
                fullWidth
              >
                {loading ? 'Updating...' : 'Update Profile'}
              </Button>
            </Grid>
          </Grid>
        </Box>
      </Paper>
    </Box>
  );
};

export default Profile;
