import { Repository, Branch, Commit, CloneRepositoryRequest, CheckoutBranchRequest, DetectChangesRequest, DetectChangesResponse, DetectCommitChangesRequest, DetectCommitChangesResponse, BuildImageRequest, BuildImageResponse, Registry, CreateRegistryRequest, UpdateRegistryRequest, DockerImage, DockerImageDetail, RetagImageRequest, CopyImageRequest } from './types';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export async function listRepositories(): Promise<Repository[]> {
  const response = await fetch(`${API_BASE_URL}/repository/list`);
  if (!response.ok) {
    throw new Error('Failed to fetch repositories');
  }
  const data = await response.json();
  return data.repositories;
}

export async function getRepositoryBranches(id: number): Promise<Branch[]> {
  try {
    // First get the repository details
    const repoResponse = await fetch(`${API_BASE_URL}/repository/${id}`);
    if (!repoResponse.ok) {
      throw new Error('Failed to fetch repository details');
    }
    const repo = await repoResponse.json();
    console.log('Repository details:', repo);

    // Then get the branches
    const branchesResponse = await fetch(`${API_BASE_URL}/imagebuilder/branches?url=${encodeURIComponent(repo.url)}`);
    if (!branchesResponse.ok) {
      throw new Error('Failed to fetch repository branches');
    }
    const data = await branchesResponse.json();
    console.log('Branches data:', data);

    if (!data.branches || !Array.isArray(data.branches)) {
      throw new Error('Invalid branches data received');
    }

    return data.branches;
  } catch (error) {
    console.error('Error in getRepositoryBranches:', error);
    throw error;
  }
}

export async function cloneRepository(request: CloneRepositoryRequest): Promise<Repository> {
  const response = await fetch(`${API_BASE_URL}/repository/clone`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(request),
  });
  if (!response.ok) {
    throw new Error('Failed to clone repository');
  }
  return response.json();
}

export async function deleteRepository(id: number): Promise<void> {
  const response = await fetch(`${API_BASE_URL}/repository/${id}`, {
    method: 'DELETE',
  });
  if (!response.ok) {
    throw new Error('Failed to delete repository');
  }
}

export const checkoutBranch = async (repositoryId: number, branchName: string): Promise<void> => {
  const response = await fetch(`${API_BASE_URL}/repository/${repositoryId}/checkout`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ branch: branchName }),
  });
  
  if (!response.ok) {
    const errorData = await response.json();
    throw new Error(errorData.error || 'Failed to checkout branch');
  }
};

export const getBranchCommits = async (repositoryId: number, branchName: string): Promise<Commit[]> => {
  const response = await fetch(`${API_BASE_URL}/repository/${repositoryId}/commits?branch=${encodeURIComponent(branchName)}`);
  if (!response.ok) {
    throw new Error('Failed to fetch branch commits');
  }
  return response.json();
};

export const syncRepository = async (repositoryId: number): Promise<void> => {
  const response = await fetch(`${API_BASE_URL}/repository/${repositoryId}/sync`, {
    method: 'POST',
  });
  
  if (!response.ok) {
    const errorData = await response.json();
    throw new Error(errorData.error || 'Failed to sync repository');
  }
};

export const detectChanges = async (request: DetectChangesRequest): Promise<DetectChangesResponse> => {
  const response = await fetch(`${API_BASE_URL}/imagebuilder/detect-changes`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(request),
  });

  if (!response.ok) {
    const errorData = await response.json();
    throw new Error(errorData.error || 'Failed to detect changes');
  }

  const data = await response.json();
  return {
    changedServices: data.changedServices || [],
  };
};

export const detectCommitChanges = async (request: DetectCommitChangesRequest): Promise<DetectCommitChangesResponse> => {
  const response = await fetch(`${API_BASE_URL}/imagebuilder/detect-commit-changes`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(request),
  });

  if (!response.ok) {
    const errorData = await response.json();
    throw new Error(errorData.error || 'Failed to detect commit changes');
  }

  const data = await response.json();
  return {
    changedServices: data.changedServices || [],
  };
};

export const buildImage = async (request: BuildImageRequest): Promise<BuildImageResponse> => {
  const response = await fetch(`${API_BASE_URL}/imagebuilder/build-multiple`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(request),
  });

  // Parse the response data
  const data = await response.json();
  
  // If the response is not OK, throw an error
  if (!response.ok) {
    throw new Error(data.error || 'Failed to build image');
  }

  // Return the parsed response data
  return {
    success: true,
    message: data.message || 'Build started successfully'
  };
};

// Registry API functions
export const listRegistries = async (): Promise<Registry[]> => {
  try {
    // Use Next.js API route instead of direct API call
    const timestamp = new Date().getTime();
    const response = await fetch(`/api/registries?_=${timestamp}`, {
      headers: {
        'Cache-Control': 'no-cache, no-store, must-revalidate',
        'Pragma': 'no-cache',
      }
    });
    
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({ error: 'Unknown error' }));
      throw new Error(errorData.error || 'Failed to fetch registries');
    }
    
    const data = await response.json();
    console.log('Raw registry data:', data);
    
    // Ensure we always return an array
    if (Array.isArray(data)) {
      return data;
    } else if (data && typeof data === 'object') {
      // If it's a single object with no array wrapper, wrap it in an array
      // Also handle the case where the data might be in a nested property
      if (data.registries && Array.isArray(data.registries)) {
        return data.registries;
      }
      return [data];
    }
    
    // Return empty array as fallback
    return [];
  } catch (error) {
    console.error('Error in listRegistries:', error);
    throw error;
  }
};

export const getRegistry = async (id: number): Promise<Registry> => {
  try {
    // Use Next.js API route instead of direct API call
    const timestamp = new Date().getTime();
    const response = await fetch(`/api/registries/${id}?_=${timestamp}`, {
      headers: {
        'Cache-Control': 'no-cache, no-store, must-revalidate',
        'Pragma': 'no-cache',
      }
    });
    
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({ error: 'Unknown error' }));
      throw new Error(errorData.error || 'Failed to fetch registry');
    }
    
    return response.json();
  } catch (error) {
    console.error('Error in getRegistry:', error);
    throw error;
  }
};

export const createRegistry = async (request: CreateRegistryRequest): Promise<Registry> => {
  try {
    // Use Next.js API route instead of direct API call
    const response = await fetch('/api/registries/create', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(request),
    });
    
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({ error: 'Unknown error' }));
      throw new Error(errorData.error || 'Failed to create registry');
    }
    
    return response.json();
  } catch (error) {
    console.error('Error in createRegistry:', error);
    throw error;
  }
};

export const updateRegistry = async (id: number, request: UpdateRegistryRequest): Promise<Registry> => {
  try {
    // Use Next.js API route instead of direct API call
    const response = await fetch(`/api/registries/${id}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(request),
    });
    
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({ error: 'Unknown error' }));
      throw new Error(errorData.error || 'Failed to update registry');
    }
    
    return response.json();
  } catch (error) {
    console.error('Error in updateRegistry:', error);
    throw error;
  }
};

export const deleteRegistry = async (id: number): Promise<void> => {
  try {
    // Use Next.js API route instead of direct API call
    const response = await fetch(`/api/registries/${id}`, {
      method: 'DELETE',
    });
    
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({ error: 'Unknown error' }));
      throw new Error(errorData.error || 'Failed to delete registry');
    }
  } catch (error) {
    console.error('Error in deleteRegistry:', error);
    throw error;
  }
};

// Docker Image Management API Functions
export const listRegistryImages = async (registryId: number): Promise<DockerImage[]> => {
  try {
    const response = await fetch(`/api/registries/${registryId}/images`, {
      headers: {
        'Cache-Control': 'no-cache, no-store, must-revalidate',
        'Pragma': 'no-cache',
      }
    });
    
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({ error: 'Unknown error' }));
      throw new Error(errorData.error || 'Failed to fetch registry images');
    }
    
    const data = await response.json();
    return data.images || [];
  } catch (error) {
    console.error('Error in listRegistryImages:', error);
    throw error;
  }
};

export const getImageDetail = async (
  registryId: number, 
  imageName: string, 
  tag: string
): Promise<DockerImageDetail> => {
  try {
    const response = await fetch(`/api/registries/${registryId}/images/${encodeURIComponent(imageName)}/${encodeURIComponent(tag)}`, {
      headers: {
        'Cache-Control': 'no-cache, no-store, must-revalidate',
        'Pragma': 'no-cache',
      }
    });
    
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({ error: 'Unknown error' }));
      throw new Error(errorData.error || 'Failed to fetch image details');
    }
    
    return response.json();
  } catch (error) {
    console.error('Error in getImageDetail:', error);
    throw error;
  }
};

export const retagImage = async (
  registryId: number, 
  request: RetagImageRequest
): Promise<void> => {
  try {
    const response = await fetch(`/api/registries/${registryId}/images/retag`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(request),
    });
    
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({ error: 'Unknown error' }));
      throw new Error(errorData.error || 'Failed to retag image');
    }
  } catch (error) {
    console.error('Error in retagImage:', error);
    throw error;
  }
};

export const deleteImage = async (
  registryId: number, 
  imageName: string, 
  tag: string
): Promise<void> => {
  try {
    // Use 'latest' as a default tag if the tag is empty
    const tagToUse = tag || 'latest';
    
    const response = await fetch(`/api/registries/${registryId}/images/${encodeURIComponent(imageName)}/${encodeURIComponent(tagToUse)}`, {
      method: 'DELETE',
    });
    
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({ error: 'Unknown error' }));
      throw new Error(errorData.error || 'Failed to delete image');
    }
  } catch (error) {
    console.error('Error in deleteImage:', error);
    throw error;
  }
};

export const copyImage = async (request: CopyImageRequest): Promise<void> => {
  try {
    const response = await fetch('/api/registries/images/copy', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(request),
    });
    
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({ error: 'Unknown error' }));
      throw new Error(errorData.error || 'Failed to copy image');
    }
  } catch (error) {
    console.error('Error in copyImage:', error);
    throw error;
  }
};

export interface TestConnectionResponse {
  status: string;
  message: string;
}

export const testRegistryConnection = async (registryId: number): Promise<TestConnectionResponse> => {
  try {
    const response = await fetch(`${API_BASE_URL}/registries/${registryId}/test-connection`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
    });
    
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({ error: 'Unknown error' }));
      throw new Error(errorData.error || 'Failed to test registry connection');
    }
    
    return response.json();
  } catch (error) {
    console.error('Error in testRegistryConnection:', error);
    throw error;
  }
}; 