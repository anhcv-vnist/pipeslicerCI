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

export interface Registry {
  id: number;
  name: string;
  url: string;
  username?: string;
  password?: string;
  description?: string;
  isOnline?: boolean;
}

export interface CreateRegistryRequest {
  name: string;
  url: string;
  username: string;
  password: string;
  description?: string;
}

export interface UpdateRegistryRequest {
  name: string;
  url: string;
  username: string;
  password: string;
  description?: string;
}

// Docker Image Types
export interface DockerImage {
  name: string;
  tags: string[];
  size: number;
  createdAt: string;
  lastUpdated: string;
}

export interface DockerImageDetail {
  name: string;
  tags: string[];
  size: number;
  createdAt: string;
  lastUpdated: string;
  layers: ImageLayer[];
  history: ImageHistory[];
  config: ImageConfig;
  labels: Record<string, string>;
}

export interface ImageLayer {
  digest: string;
  size: number;
  createdAt: string;
}

export interface ImageHistory {
  created: string;
  createdBy: string;
  comment: string;
  emptyLayer: boolean;
}

export interface ImageConfig {
  architecture: string;
  os: string;
  env: string[];
  labels: Record<string, string>;
}

export interface RetagImageRequest {
  source_image: string;
  source_tag: string;
  destination_image: string;
  destination_tag: string;
}

export interface CopyImageRequest {
  sourceRegistryId: number;
  sourceImage: string;
  sourceTag: string;
  destinationRegistryId: number;
  destinationImage: string;
  destinationTag: string;
} 