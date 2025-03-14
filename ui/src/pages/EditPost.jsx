import React, { useState, useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useNavigate, useParams } from 'react-router-dom';
import {
  Box,
  TextField,
  Button,
  Typography,
  Paper,
  Chip,
  IconButton,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Alert,
} from '@mui/material';
import { Add as AddIcon, Close as CloseIcon } from '@mui/icons-material';
import { fetchPostById, updatePost } from '../store/slices/postSlice';
import { uploadMedia } from '../store/slices/mediaSlice';
import ImageUpload from '../components/ImageUpload';

const EditPost = () => {
  const { id } = useParams();
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const { currentPost, loading, error } = useSelector((state) => state.posts);

  const [formData, setFormData] = useState({
    title: '',
    content: '',
    status: 'draft',
    tags: [],
    featured_image: null,
    gallery: [],
  });
  const [tagInput, setTagInput] = useState('');

  useEffect(() => {
    dispatch(fetchPostById(id));
  }, [dispatch, id]);

  useEffect(() => {
    if (currentPost) {
      setFormData({
        title: currentPost.title || '',
        content: currentPost.content || '',
        status: currentPost.status || 'draft',
        tags: currentPost.tags || [],
        featured_image: currentPost.featured_image || null,
        gallery: currentPost.gallery || [],
      });
    }
  }, [currentPost]);

  const handleChange = (e) => {
    setFormData({ ...formData, [e.target.name]: e.target.value });
  };

  const handleTagInputChange = (e) => {
    setTagInput(e.target.value);
  };

  const handleAddTag = () => {
    if (tagInput && !formData.tags.includes(tagInput)) {
      setFormData({
        ...formData,
        tags: [...formData.tags, tagInput],
      });
      setTagInput('');
    }
  };

  const handleRemoveTag = (tagToRemove) => {
    setFormData({
      ...formData,
      tags: formData.tags.filter((tag) => tag !== tagToRemove),
    });
  };

  const handleFeaturedImageUpload = async (file) => {
    try {
      const formData = new FormData();
      formData.append('file', file);
      const response = await dispatch(uploadMedia(formData)).unwrap();
      setFormData((prev) => ({
        ...prev,
        featured_image: response,
      }));
    } catch (error) {
      console.error('Failed to upload image:', error);
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      await dispatch(updatePost({ id, postData: formData })).unwrap();
      navigate(`/posts/${id}`);
    } catch (error) {
      console.error('Failed to update post:', error);
    }
  };

  if (loading) {
    return (
      <Typography variant="h6" align="center">
        Loading...
      </Typography>
    );
  }

  if (!currentPost) {
    return (
      <Typography variant="h6" align="center" color="error">
        Post not found
      </Typography>
    );
  }

  return (
    <Paper sx={{ p: 4, maxWidth: 800, mx: 'auto' }}>
      <Typography variant="h4" gutterBottom>
        Edit Post
      </Typography>

      {error && (
        <Alert severity="error" sx={{ mb: 3 }}>
          {error}
        </Alert>
      )}

      <Box component="form" onSubmit={handleSubmit}>
        <TextField
          fullWidth
          label="Title"
          name="title"
          value={formData.title}
          onChange={handleChange}
          required
          sx={{ mb: 3 }}
        />

        <TextField
          fullWidth
          label="Content"
          name="content"
          value={formData.content}
          onChange={handleChange}
          required
          multiline
          rows={6}
          sx={{ mb: 3 }}
        />

        <FormControl fullWidth sx={{ mb: 3 }}>
          <InputLabel>Status</InputLabel>
          <Select
            name="status"
            value={formData.status}
            onChange={handleChange}
            label="Status"
          >
            <MenuItem value="draft">Draft</MenuItem>
            <MenuItem value="published">Published</MenuItem>
          </Select>
        </FormControl>

        <Box sx={{ mb: 3 }}>
          <Typography variant="subtitle1" gutterBottom>
            Featured Image
          </Typography>
          <ImageUpload
            onUpload={handleFeaturedImageUpload}
            preview={formData.featured_image?.path}
          />
        </Box>

        <Box sx={{ mb: 3 }}>
          <Typography variant="subtitle1" gutterBottom>
            Tags
          </Typography>
          <Box sx={{ display: 'flex', gap: 1, mb: 1 }}>
            <TextField
              size="small"
              value={tagInput}
              onChange={handleTagInputChange}
              placeholder="Add a tag"
            />
            <IconButton onClick={handleAddTag} color="primary">
              <AddIcon />
            </IconButton>
          </Box>
          <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1 }}>
            {formData.tags.map((tag) => (
              <Chip
                key={tag}
                label={tag}
                onDelete={() => handleRemoveTag(tag)}
                deleteIcon={<CloseIcon />}
              />
            ))}
          </Box>
        </Box>

        <Button
          type="submit"
          variant="contained"
          color="primary"
          disabled={loading}
          fullWidth
        >
          {loading ? 'Updating...' : 'Update Post'}
        </Button>
      </Box>
    </Paper>
  );
};

export default EditPost;
