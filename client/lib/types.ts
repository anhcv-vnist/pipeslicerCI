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