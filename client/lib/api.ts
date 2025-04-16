import { Repository, Branch, Commit, CloneRepositoryRequest, CheckoutBranchRequest } from './types';

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
  const repo = await fetch(`${API_BASE_URL}/repository/${id}`).then(res => res.json());
  const response = await fetch(`${API_BASE_URL}/imagebuilder/branches?url=${encodeURIComponent(repo.url)}`);
  if (!response.ok) {
    throw new Error('Failed to fetch repository branches');
  }
  const data = await response.json();
  return data.branches;
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