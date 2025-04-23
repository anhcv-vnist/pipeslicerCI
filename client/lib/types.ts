export interface Repository {
  id: number;
  url: string;
  name: string;
  description: string;
  localPath: string;
  lastUpdated: string;
  createdAt: string;
  updatedAt: string;
  currentBranch?: string;
  branches?: Branch[];
}

export interface Branch {
  name: string;
  isCurrent: boolean;
  lastCommit: string;
  lastCommitDate: string;
  isRemote: boolean;
  remoteName?: string;
}

export interface Commit {
  hash: string;
  message: string;
  author: string;
  timestamp: string;
}

export interface CloneRepositoryRequest {
  url: string;
  name: string;
  description: string;
}

export interface CheckoutBranchRequest {
  branch: string;
}

export interface DetectChangesRequest {
  url: string;
  baseBranch: string;
  currentBranch: string;
}

export interface ChangedService {
  path: string;
  hasDockerfile: boolean;
}

export interface DetectChangesResponse {
  changedServices: ChangedService[];
}

export interface DetectCommitChangesRequest {
  url: string;
  baseCommit: string;
  currentCommit: string;
}

export interface DetectCommitChangesResponse {
  changedServices: ChangedService[];
}

export interface BuildImageRequest {
  url: string;
  branch: string;
  servicePaths: string[];
  tag: string;
  registry: string;
}

export interface BuildImageResponse {
  success: boolean;
  message: string;
} 