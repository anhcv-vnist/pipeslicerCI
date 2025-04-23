import { Repository, Branch, Commit, CloneRepositoryRequest, CheckoutBranchRequest, DetectChangesRequest, DetectChangesResponse, DetectCommitChangesRequest, DetectCommitChangesResponse, BuildImageRequest, BuildImageResponse } from './types';

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