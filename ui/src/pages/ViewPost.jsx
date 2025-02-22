import React, { useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useParams, useNavigate, Link as RouterLink } from 'react-router-dom';
import {
  Box,
  Typography,
  Paper,
  Chip,
  Button,
  Grid,
  Divider,
  Avatar,
  IconButton,
} from '@mui/material';
import {
  Edit as EditIcon,
  Delete as DeleteIcon,
  ArrowBack as ArrowBackIcon,
} from '@mui/icons-material';
import { fetchPostById, deletePost } from '../store/slices/postSlice';
import config from '../config';

const ViewPost = () => {
  const { id } = useParams();
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const { currentPost, loading, error } = useSelector((state) => state.posts);
  const { user } = useSelector((state) => state.auth);

  useEffect(() => {
    dispatch(fetchPostById(id));
  }, [dispatch, id]);

  const handleDelete = async () => {
    if (window.confirm('Are you sure you want to delete this post?')) {
      try {
        await dispatch(deletePost(id)).unwrap();
        navigate('/');
      } catch (error) {
        console.error('Failed to delete post:', error);
      }
    }
  };

  if (loading) {
    return (
      <Typography variant="h6" align="center">
        Loading...
      </Typography>
    );
  }

  if (error) {
    return (
      <Typography variant="h6" align="center" color="error">
        {error}
      </Typography>
    );
  }

  if (!currentPost) {
    return (
      <Typography variant="h6" align="center">
        Post not found
      </Typography>
    );
  }

  const isAuthor = user && user.id === currentPost.author_id;

  return (
    <Box maxWidth="lg" sx={{ mx: 'auto', py: 4 }}>
      <Button
        component={RouterLink}
        to="/"
        startIcon={<ArrowBackIcon />}
        sx={{ mb: 2 }}
      >
        Back to Posts
      </Button>

      <Paper sx={{ p: 4 }}>
        {currentPost.featured_image && (
          <Box
            component="img"
            src={`${config.mediaUrl}/${currentPost.featured_image.path}`}
            alt={currentPost.title}
            sx={{
              width: '100%',
              height: 400,
              objectFit: 'cover',
              borderRadius: 1,
              mb: 4,
            }}
          />
        )}

        <Grid container spacing={2} alignItems="center" sx={{ mb: 4 }}>
          <Grid item>
            <Avatar
              src={currentPost.author?.profile?.avatar}
              alt={currentPost.author?.username}
            />
          </Grid>
          <Grid item xs>
            <Typography variant="subtitle1">
              {currentPost.author?.username}
            </Typography>
            <Typography variant="body2" color="text.secondary">
              {new Date(currentPost.created_at).toLocaleDateString()}
            </Typography>
          </Grid>
          {isAuthor && (
            <Grid item>
              <IconButton
                component={RouterLink}
                to={`/posts/${id}/edit`}
                color="primary"
              >
                <EditIcon />
              </IconButton>
              <IconButton onClick={handleDelete} color="error">
                <DeleteIcon />
              </IconButton>
            </Grid>
          )}
        </Grid>

        <Typography variant="h4" gutterBottom>
          {currentPost.title}
        </Typography>

        <Box sx={{ mb: 3 }}>
          {currentPost.tags?.map((tag) => (
            <Chip
              key={tag}
              label={tag}
              variant="outlined"
              size="small"
              sx={{ mr: 1 }}
            />
          ))}
        </Box>

        <Divider sx={{ my: 3 }} />

        <Typography variant="body1" sx={{ whiteSpace: 'pre-wrap' }}>
          {currentPost.content}
        </Typography>

        {currentPost.gallery?.length > 0 && (
          <Box sx={{ mt: 4 }}>
            <Typography variant="h6" gutterBottom>
              Gallery
            </Typography>
            <Grid container spacing={2}>
              {currentPost.gallery.map((image) => (
                <Grid item xs={12} sm={6} md={4} key={image.id}>
                  <Box
                    component="img"
                    src={`${config.mediaUrl}/${image.path}`}
                    alt={`Gallery image ${image.id}`}
                    sx={{
                      width: '100%',
                      height: 200,
                      objectFit: 'cover',
                      borderRadius: 1,
                    }}
                  />
                </Grid>
              ))}
            </Grid>
          </Box>
        )}
      </Paper>
    </Box>
  );
};

export default ViewPost;
