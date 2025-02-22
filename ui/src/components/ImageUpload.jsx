import React, { useRef } from 'react';
import { Box, Button, Typography } from '@mui/material';
import { CloudUpload as CloudUploadIcon } from '@mui/icons-material';

const ImageUpload = ({ onUpload, preview, multiple = false }) => {
  const fileInputRef = useRef();

  const handleClick = () => {
    fileInputRef.current.click();
  };

  const handleChange = (event) => {
    const files = event.target.files;
    if (files.length > 0) {
      if (multiple) {
        onUpload(Array.from(files));
      } else {
        onUpload(files[0]);
      }
    }
  };

  return (
    <Box>
      <input
        type="file"
        ref={fileInputRef}
        onChange={handleChange}
        accept="image/*"
        multiple={multiple}
        style={{ display: 'none' }}
      />
      <Box
        sx={{
          border: '2px dashed #ccc',
          borderRadius: 2,
          p: 2,
          textAlign: 'center',
          cursor: 'pointer',
          '&:hover': {
            borderColor: 'primary.main',
          },
        }}
        onClick={handleClick}
      >
        {preview ? (
          <Box
            component="img"
            src={preview}
            alt="Preview"
            sx={{
              maxWidth: '100%',
              maxHeight: 200,
              objectFit: 'contain',
            }}
          />
        ) : (
          <>
            <CloudUploadIcon sx={{ fontSize: 48, color: 'text.secondary' }} />
            <Typography variant="body1" color="text.secondary">
              Click or drag image to upload
            </Typography>
          </>
        )}
      </Box>
      <Button
        variant="outlined"
        onClick={handleClick}
        startIcon={<CloudUploadIcon />}
        sx={{ mt: 2 }}
      >
        Choose Image
      </Button>
    </Box>
  );
};

export default ImageUpload;
