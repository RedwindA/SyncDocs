import axios from 'axios';

// Create an Axios instance
const apiClient = axios.create({
  baseURL: '/api', // All API requests will be prefixed with /api
  timeout: 10000, // Request timeout: 10 seconds
  headers: {
    'Content-Type': 'application/json',
  },
  // IMPORTANT: Basic Auth credentials need to be handled securely.
  // Storing them directly in code is NOT recommended for production.
  // For this example, we'll assume the browser's built-in auth prompt
  // or a login mechanism handles setting the Authorization header.
  // If you need programmatic Basic Auth (e.g., from stored credentials),
  // you would set the 'auth' property or 'Authorization' header here,
  // but be mindful of security implications.
  // Example (use with caution):
  // auth: {
  //   username: 'YOUR_USERNAME', // Ideally from a secure source
  //   password: 'YOUR_PASSWORD', // Ideally from a secure source
  // },
});

// Add a response interceptor to handle errors globally
apiClient.interceptors.response.use(
  (response) => response, // Simply return successful responses
  (error) => {
    // Handle specific error statuses
    if (error.response) {
      console.error('API Error:', error.response.status, error.response.data);
      if (error.response.status === 401) {
        // Handle unauthorized access, e.g., redirect to login or show message
        // This might happen if Basic Auth fails or session expires
        alert('Unauthorized access. Please check your credentials or log in again.');
        // Potentially clear stored credentials and redirect
        // window.location.href = '/login'; // If you have a login route
      } else if (error.response.status === 404) {
        // Handle not found errors
        // alert('Resource not found.');
      } else {
        // Handle other errors (500, 400, etc.)
        const errorMsg = error.response.data?.error || 'An unexpected error occurred.';
        alert(`Error: ${errorMsg}`);
      }
    } else if (error.request) {
      // The request was made but no response was received
      console.error('Network Error:', error.request);
      alert('Network error. Please check your connection.');
    } else {
      // Something happened in setting up the request that triggered an Error
      console.error('Error:', error.message);
      alert('An error occurred while setting up the request.');
    }
    // Return a rejected promise to propagate the error
    return Promise.reject(error);
  }
);

// Define interfaces for API data (mirroring backend models)
// We can refine these as needed
export interface RepositoryListItem {
  id: number;
  url: string;
  docs_path: string;
  extensions: string;
  last_sync_status: string;
  // Update last_sync_time to match the actual JSON structure from sql.NullTime
  last_sync_time: { Time: string; Valid: boolean; } | null;
  last_sync_error: string;
  updated_at: string; // ISO string
}

export interface Repository extends RepositoryListItem {
    owner: string;
    repo_name: string;
    aggregated_content: string | null;
    created_at: string; // ISO string
}


export interface RepositoryCreatePayload {
  url: string;
  docs_path: string;
  extensions: string;
}

export interface RepositoryUpdatePayload {
  docs_path: string;
  extensions: string;
}


// Define API functions
const apiService = {
  listRepositories(): Promise<RepositoryListItem[]> {
    return apiClient.get('/repositories').then(response => response.data);
  },

  getRepository(id: number): Promise<Repository> {
    return apiClient.get(`/repositories/${id}`).then(response => response.data);
  },

  createRepository(payload: RepositoryCreatePayload): Promise<RepositoryListItem> {
    return apiClient.post('/repositories', payload).then(response => response.data);
  },

  updateRepository(id: number, payload: RepositoryUpdatePayload): Promise<RepositoryListItem> {
    return apiClient.put(`/repositories/${id}`, payload).then(response => response.data);
  },

  deleteRepository(id: number): Promise<{ message: string }> {
    return apiClient.delete(`/repositories/${id}`).then(response => response.data);
  },

  triggerSync(id: number): Promise<{ message: string }> {
    return apiClient.post(`/repositories/${id}/sync`).then(response => response.data);
  },

  // Note: Downloading is typically handled via a direct link or window.location,
  // as Axios isn't ideal for triggering file downloads directly in the browser
  // in a user-friendly way without extra steps.
  getDownloadUrl(id: number): string {
    // We return the URL, and the component can use it in an <a> tag
    // or window.location.href
    return `/api/repositories/${id}/download`;
  }
};

export default apiService;
