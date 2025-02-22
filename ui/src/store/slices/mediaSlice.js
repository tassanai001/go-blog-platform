import { createSlice, createAsyncThunk } from '@reduxjs/toolkit';
import axios from 'axios';

const API_URL = 'http://localhost:8080/api';

// Create axios instance with auth header
const api = axios.create({
  baseURL: API_URL,
});

// Add auth token to requests
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

export const uploadMedia = createAsyncThunk(
  'media/uploadMedia',
  async (formData, { rejectWithValue }) => {
    try {
      const response = await api.post('/media', formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      });
      return response.data;
    } catch (error) {
      return rejectWithValue(error.response.data);
    }
  }
);

export const deleteMedia = createAsyncThunk(
  'media/deleteMedia',
  async (id, { rejectWithValue }) => {
    try {
      await api.delete(`/media/${id}`);
      return id;
    } catch (error) {
      return rejectWithValue(error.response.data);
    }
  }
);

const initialState = {
  uploads: [],
  loading: false,
  error: null,
};

const mediaSlice = createSlice({
  name: 'media',
  initialState,
  reducers: {
    clearError: (state) => {
      state.error = null;
    },
    clearUploads: (state) => {
      state.uploads = [];
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(uploadMedia.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(uploadMedia.fulfilled, (state, action) => {
        state.loading = false;
        state.uploads.push(action.payload);
      })
      .addCase(uploadMedia.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload?.error || 'Failed to upload media';
      })
      .addCase(deleteMedia.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(deleteMedia.fulfilled, (state, action) => {
        state.loading = false;
        state.uploads = state.uploads.filter(
          (media) => media.id !== action.payload
        );
      })
      .addCase(deleteMedia.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload?.error || 'Failed to delete media';
      });
  },
});

export const { clearError, clearUploads } = mediaSlice.actions;
export default mediaSlice.reducer;
