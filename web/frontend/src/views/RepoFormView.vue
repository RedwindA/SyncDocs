<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
// Remove unused 'Repository' type import, keep others
import apiService, { type RepositoryCreatePayload, type RepositoryUpdatePayload } from '../services/api';

const route = useRoute();
const router = useRouter();

const formData = ref<RepositoryCreatePayload | RepositoryUpdatePayload>({
	url: '', // Only for create mode
	docs_path: '',
	extensions: 'md,mdx', // Default value
	branch: '', // Only for create mode, add branch
});

const isLoading = ref(false);
const errorMessage = ref<string | null>(null);
const pageTitle = ref('Add New Repository');

// Determine if we are in edit mode based on route params
const isEditMode = computed(() => !!route.params.id);
const repoId = computed(() => Number(route.params.id));

async function fetchRepoDataForEdit() {
  if (!isEditMode.value || isNaN(repoId.value)) return;

  isLoading.value = true;
  errorMessage.value = null;
  pageTitle.value = `Edit Repository ID: ${repoId.value}`;
  try {
    // Fetch the full repo details to populate the form
    const repo = await apiService.getRepository(repoId.value);
    // Populate form data for editing (URL is not editable)
    // Populate form data for editing (URL and branch are not editable here)
    formData.value = {
    	docs_path: repo.docs_path,
    	extensions: repo.extensions,
    	// branch is not part of RepositoryUpdatePayload, so it's fine
    };
    } catch (error: any) {
    console.error(`Failed to fetch repository details for editing (ID ${repoId.value}):`, error);
     if (error.response && error.response.status === 404) {
        errorMessage.value = `Repository with ID ${repoId.value} not found for editing.`;
    } else {
        errorMessage.value = 'Failed to load repository data for editing.';
    }
  } finally {
    isLoading.value = false;
  }
}

async function handleSubmit() {
  isLoading.value = true;
  errorMessage.value = null;

  try {
    if (isEditMode.value) {
      // Update existing repository
      await apiService.updateRepository(repoId.value, formData.value as RepositoryUpdatePayload);
      alert(`Repository ${repoId.value} updated successfully!`);
      router.push({ name: 'Home' }); // Redirect back to list
    } else {
      // Create new repository
      await apiService.createRepository(formData.value as RepositoryCreatePayload);
      alert('Repository added successfully!');
      router.push({ name: 'Home' }); // Redirect back to list
    }
  } catch (error: any) {
    console.error('Failed to save repository:', error);
    // Error message is usually handled by the interceptor, but we can set one here if needed
    errorMessage.value = error.response?.data?.error || 'Failed to save repository. Please check the details and try again.';
  } finally {
    isLoading.value = false;
  }
}

function goBack() {
  router.push({ name: 'Home' });
}

// Fetch data when the component mounts if in edit mode
onMounted(() => {
  if (isEditMode.value) {
    fetchRepoDataForEdit();
  }
});

// Watch for route changes (e.g., navigating from add to edit)
// This might not be strictly necessary depending on app structure, but can be useful
watch(() => route.params.id, (newId, oldId) => {
    if (newId && newId !== oldId && isEditMode.value) {
        fetchRepoDataForEdit();
    } else if (!newId) {
        // Reset form if navigating back to 'new' mode (or handle differently)
        formData.value = { url: '', docs_path: '', extensions: 'md,mdx', branch: '' };
        pageTitle.value = 'Add New Repository';
        errorMessage.value = null;
    }
});

</script>

<template>
  <div class="repo-form-view">
    <button @click="goBack" class="btn btn-secondary back-btn">&larr; Back to List</button>

    <h1>{{ pageTitle }}</h1>

    <div v-if="isLoading && isEditMode" class="loading">Loading repository data...</div>
    <div v-if="errorMessage" class="error-message">{{ errorMessage }}</div>

    <form @submit.prevent="handleSubmit" v-if="!isLoading || !isEditMode">
      <div class="form-group" v-if="!isEditMode">
        <label for="repo-url">GitHub Repository URL:</label>
        <input
          type="url"
          id="repo-url"
          v-model="(formData as RepositoryCreatePayload).url"
          placeholder="https://github.com/owner/repo"
          required
          :disabled="isLoading"
        />
        <small>Example: https://github.com/vuejs/docs</small>
         </div>
      
         <div class="form-group" v-if="!isEditMode">
        <label for="repo-branch">Branch (Optional):</label>
        <input
          type="text"
          id="repo-branch"
          v-model="(formData as RepositoryCreatePayload).branch"
          placeholder="e.g., main, develop"
          :disabled="isLoading"
        />
        <small>Leave empty to use the default branch of the repository.</small>
         </div>
      
          <div class="form-group" v-if="isEditMode">
        <label>GitHub Repository URL:</label>
        <p><i>URL cannot be changed after creation.</i></p>
      </div>

      <div class="form-group">
        <label for="docs-path">Documentation Path:</label>
        <input
          type="text"
          id="docs-path"
          v-model="formData.docs_path"
          placeholder="e.g., docs, content/blog, src/pages"
          required
          :disabled="isLoading"
        />
        <small>Path relative to the repository root.</small>
      </div>

      <div class="form-group">
        <label for="extensions">File Extensions (comma-separated):</label>
        <input
          type="text"
          id="extensions"
          v-model="formData.extensions"
          placeholder="e.g., md,mdx,txt"
          required
          :disabled="isLoading"
        />
         <small>Only files with these extensions will be synced.</small>
      </div>

      <div class="form-actions">
        <button type="submit" class="btn btn-primary" :disabled="isLoading">
          {{ isLoading ? 'Saving...' : (isEditMode ? 'Update Repository' : 'Add Repository') }}
        </button>
        <button type="button" @click="goBack" class="btn btn-secondary" :disabled="isLoading">
          Cancel
        </button>
      </div>
    </form>
  </div>
</template>

<style scoped>
.repo-form-view {
  padding: 20px;
  max-width: 600px;
  margin: 0 auto;
  font-family: sans-serif;
}

.back-btn {
  margin-bottom: 20px;
}

h1 {
    margin-bottom: 25px;
    text-align: center;
}

.loading, .error-message {
  margin-bottom: 20px; /* Adjusted margin */
  padding: 15px;
  border-radius: 4px;
}

.loading {
  background-color: #e0e0e0;
}

.error-message {
  background-color: #f8d7da;
  color: #721c24;
  border: 1px solid #f5c6cb;
}


.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 5px;
  font-weight: bold;
}

.form-group input[type="text"],
.form-group input[type="url"] {
  width: 100%;
  padding: 10px;
  border: 1px solid #ccc;
  border-radius: 4px;
  box-sizing: border-box; /* Include padding and border in element's total width/height */
}

.form-group input:disabled {
    background-color: #e9ecef;
    opacity: 1;
}

.form-group small {
    display: block;
    margin-top: 5px;
    font-size: 0.85em;
    color: #6c757d;
}

.form-group p {
    margin: 5px 0;
    font-style: italic;
    color: #6c757d;
}

.form-actions {
  margin-top: 30px;
  display: flex;
  gap: 10px; /* Add space between buttons */
}

/* Basic Button Styling */
.btn {
  display: inline-block; font-weight: 400; text-align: center; vertical-align: middle; user-select: none; border: 1px solid transparent; padding: 0.375rem 0.75rem; font-size: 1rem; line-height: 1.5; border-radius: 0.25rem; transition: color 0.15s ease-in-out, background-color 0.15s ease-in-out, border-color 0.15s ease-in-out, box-shadow 0.15s ease-in-out; cursor: pointer; text-decoration: none;
}
.btn:disabled { cursor: not-allowed; opacity: 0.65; }
.btn-primary { color: #fff; background-color: #007bff; border-color: #007bff; }
.btn-primary:hover:not(:disabled) { background-color: #0056b3; border-color: #0056b3; }
.btn-secondary { color: #fff; background-color: #6c757d; border-color: #6c757d; }
.btn-secondary:hover:not(:disabled) { background-color: #5a6268; border-color: #545b62; }
</style>
