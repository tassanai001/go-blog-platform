import React, { useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Link as RouterLink } from 'react-router-dom';
import {
  Grid,
  Card,
  CardContent,
  CardMedia,
  Typography,
  Button,
  Box,
  Chip,
  Skeleton,
} from '@mui/material';
import { fetchPosts } from '../store/slices/postSlice';

const Home = () => {
  const dispatch = useDispatch();
  const { posts, loading } = useSelector((state) => state.posts);

  useEffect(() => {
    dispatch(fetchPosts());
  }, [dispatch]);

  if (loading) {
    return (
      <Grid container spacing={4}>
        {[1, 2, 3].map((n) => (
          <Grid item xs={12} md={4} key={n}>
            <Card>
              <Skeleton variant="rectangular" height={200} />
              <CardContent>
                <Skeleton variant="text" height={40} />
                <Skeleton variant="text" height={20} />
                <Skeleton variant="text" height={20} />
              </CardContent>
            </Card>
          </Grid>
        ))}
      </Grid>
    );
  }

  return (
    <Box>
      <Typography variant="h4" gutterBottom>
        Latest Posts
      </Typography>
      <Grid container spacing={4}>
        {posts.map((post) => (
          <Grid item xs={12} md={4} key={post.id}>
            <Card>
              {post.featured_image && (
                <CardMedia
                  component="img"
                  height="200"
                  image={post.featured_image.path}
                  alt={post.title}
                />
              )}
              <CardContent>
                <Typography variant="h6" gutterBottom>
                  {post.title}
                </Typography>
                <Typography
                  variant="body2"
                  color="text.secondary"
                  sx={{
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    display: '-webkit-box',
                    WebkitLineClamp: 3,
                    WebkitBoxOrient: 'vertical',
                    mb: 2,
                  }}
                >
                  {post.content}
                </Typography>
                <Box sx={{ mb: 2 }}>
                  {post.tags.map((tag) => (
                    <Chip
                      key={tag}
                      label={tag}
                      size="small"
                      sx={{ mr: 0.5, mb: 0.5 }}
                    />
                  ))}
                </Box>
                <Button
                  component={RouterLink}
                  to={`/posts/${post.id}`}
                  variant="contained"
                  color="primary"
                >
                  Read More
                </Button>
              </CardContent>
            </Card>
          </Grid>
        ))}
      </Grid>
    </Box>
  );
};

export default Home;
