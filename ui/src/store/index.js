import { configureStore } from '@reduxjs/toolkit';
import authReducer from './slices/authSlice';
import postReducer from './slices/postSlice';
import mediaReducer from './slices/mediaSlice';

const store = configureStore({
  reducer: {
    auth: authReducer,
    posts: postReducer,
    media: mediaReducer,
  },
});

export default store;
