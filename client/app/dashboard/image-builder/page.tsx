'use client';

import { useState, useEffect } from 'react';
import { Repository, Branch, ChangedService, Commit } from '@/lib/types';
import { listRepositories, getRepositoryBranches, detectChanges, detectCommitChanges, buildImage, getBranchCommits } from '@/lib/api';
import { Button } from '@/components/ui/button';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useToast } from '@/hooks/use-toast';
import { GitBranch, ArrowRight, ChevronDown, Globe, Laptop, AlertCircle, Loader2, GitCompare, GitCommit, X, Package, CheckSquare, Square } from 'lucide-react';
import { Input } from "@/components/ui/input";
import { Checkbox } from "@/components/ui/checkbox";
import { cn } from "@/lib/utils";

const BRANCHES_PER_PAGE = 20;

// Helper function to safely access localStorage
const getLocalStorageItem = (key: string): string | null => {
  if (typeof window !== 'undefined') {
    return localStorage.getItem(key);
  }
  return null;
};

// Helper function to safely set localStorage
const setLocalStorageItem = (key: string, value: string): void => {
  if (typeof window !== 'undefined') {
    localStorage.setItem(key, value);
  }
};

// Helper function to safely remove localStorage
const removeLocalStorageItem = (key: string): void => {
  if (typeof window !== 'undefined') {
    localStorage.removeItem(key);
  }
};

// Move BranchSelect component outside of the main component
const BranchSelect = ({ 
  value, 
  onValueChange, 
  placeholder, 
  excludeBranch,
  displayedLocalBranches,
  displayedRemoteBranches,
  allBranches,
  loadingBranches,
  onShowMoreLocal,
  onShowMoreRemote,
  disabled
}: { 
  value: string, 
  onValueChange: (value: string) => void,
  placeholder: string,
  excludeBranch?: string,
  displayedLocalBranches: Branch[],
  displayedRemoteBranches: Branch[],
  allBranches: Branch[],
  loadingBranches: boolean,
  onShowMoreLocal: () => void,
  onShowMoreRemote: () => void,
  disabled: boolean
}) => {
  const getFilteredBranches = (branches: Branch[], excludeBranch?: string) => {
    return branches.filter(branch => branch.name !== excludeBranch);
  };

  const filteredLocalBranches = getFilteredBranches(displayedLocalBranches, excludeBranch);
  const filteredRemoteBranches = getFilteredBranches(displayedRemoteBranches, excludeBranch);

  // Sort branches alphabetically, but prioritize current branch
  const sortBranches = (branches: Branch[]) => {
    return [...branches].sort((a, b) => {
      if (a.isCurrent) return -1;
      if (b.isCurrent) return 1;
      return a.name.localeCompare(b.name);
    });
  };

  const sortedLocalBranches = sortBranches(filteredLocalBranches);
  const sortedRemoteBranches = sortBranches(filteredRemoteBranches);

  // Calculate remaining counts
  const remainingLocalBranches = getFilteredBranches(allBranches.filter(b => !b.isRemote), excludeBranch).length - filteredLocalBranches.length;
  const remainingRemoteBranches = getFilteredBranches(allBranches.filter(b => b.isRemote), excludeBranch).length - filteredRemoteBranches.length;

  return (
    <Select
      value={value}
      onValueChange={onValueChange}
      disabled={loadingBranches || disabled}
    >
      <SelectTrigger>
        <SelectValue placeholder={loadingBranches ? "Loading branches..." : placeholder} />
      </SelectTrigger>
      <SelectContent>
        {/* Local Branches */}
        {sortedLocalBranches.length > 0 && (
          <>
            <div className="px-2 py-1.5 text-sm font-semibold text-muted-foreground flex items-center gap-2">
              <Laptop className="h-4 w-4" />
              Local Branches
            </div>
            {sortedLocalBranches.map((branch) => (
              <SelectItem key={`local-${branch.name}`} value={branch.name}>
                <div className="flex items-center gap-2">
                  <GitBranch className="h-4 w-4" />
                  {branch.name}
                  {branch.isCurrent && (
                    <span className="text-xs px-2 py-1 rounded-full bg-primary/10 text-primary">
                      Current
                    </span>
                  )}
                </div>
              </SelectItem>
            ))}
            {remainingLocalBranches > 0 && (
              <div className="p-2">
                <Button
                  variant="ghost"
                  className="w-full justify-center"
                  onClick={onShowMoreLocal}
                >
                  <ChevronDown className="h-4 w-4 mr-2" />
                  Show More Local ({remainingLocalBranches} remaining)
                </Button>
              </div>
            )}
          </>
        )}

        {/* Remote Branches */}
        {sortedRemoteBranches.length > 0 && (
          <>
            <div className="px-2 py-1.5 text-sm font-semibold text-muted-foreground flex items-center gap-2 mt-2">
              <Globe className="h-4 w-4" />
              Remote Branches
            </div>
            {sortedRemoteBranches.map((branch) => (
              <SelectItem key={`remote-${branch.name}`} value={branch.name}>
                <div className="flex items-center gap-2">
                  <GitBranch className="h-4 w-4" />
                  {branch.name}
                  <span className="text-xs text-muted-foreground">
                    (remote)
                  </span>
                </div>
              </SelectItem>
            ))}
            {remainingRemoteBranches > 0 && (
              <div className="p-2">
                <Button
                  variant="ghost"
                  className="w-full justify-center"
                  onClick={onShowMoreRemote}
                >
                  <ChevronDown className="h-4 w-4 mr-2" />
                  Show More Remote ({remainingRemoteBranches} remaining)
                </Button>
              </div>
            )}
          </>
        )}
      </SelectContent>
    </Select>
  );
};

export default function ImageBuilderPage() {
  const [repositories, setRepositories] = useState<Repository[]>([]);
  const [selectedRepo, setSelectedRepo] = useState<Repository | null>(null);
  const [allBranches, setAllBranches] = useState<Branch[]>([]);
  const [displayedLocalBranches, setDisplayedLocalBranches] = useState<Branch[]>([]);
  const [displayedRemoteBranches, setDisplayedRemoteBranches] = useState<Branch[]>([]);
  const [baseBranch, setBaseBranch] = useState<string>('');
  const [currentBranch, setCurrentBranch] = useState<string>('');
  const [baseCommit, setBaseCommit] = useState<string>('');
  const [currentCommit, setCurrentCommit] = useState<string>('');
  const [changedServices, setChangedServices] = useState<ChangedService[] | null>(null);
  const [loading, setLoading] = useState(false);
  const [loadingBranches, setLoadingBranches] = useState(false);
  const [activeComparison, setActiveComparison] = useState<'branch' | 'commit' | null>(null);
  const [branchValidationAttempted, setBranchValidationAttempted] = useState(false);
  const [commitValidationAttempted, setCommitValidationAttempted] = useState(false);
  const [isClient, setIsClient] = useState(false);
  const { toast } = useToast();
  const [buildingServices, setBuildingServices] = useState<{[key: string]: boolean}>({});
  const [buildTag, setBuildTag] = useState<string>('v1.0.0');
  const [registry, setRegistry] = useState<string>('localhost:5000');
  const [selectedServices, setSelectedServices] = useState<string[]>([]);
  const [isBuilding, setIsBuilding] = useState(false);
  const [baseBranchForCommits, setBaseBranchForCommits] = useState<string>('');
  const [currentBranchForCommits, setCurrentBranchForCommits] = useState<string>('');
  const [baseCommits, setBaseCommits] = useState<Commit[]>([]);
  const [currentCommits, setCurrentCommits] = useState<Commit[]>([]);
  const [loadingBaseCommits, setLoadingBaseCommits] = useState(false);
  const [loadingCurrentCommits, setLoadingCurrentCommits] = useState(false);
  const [displayedBaseCommits, setDisplayedBaseCommits] = useState<Commit[]>([]);
  const [displayedCurrentCommits, setDisplayedCurrentCommits] = useState<Commit[]>([]);
  const [baseCommitsPage, setBaseCommitsPage] = useState(1);
  const [currentCommitsPage, setCurrentCommitsPage] = useState(1);
  const COMMITS_PER_PAGE = 10;

  // Load saved state after component mounts
  useEffect(() => {
    setIsClient(true);
    const savedRepo = getLocalStorageItem('selectedRepo');
    if (savedRepo) {
      setSelectedRepo(JSON.parse(savedRepo));
    }
    setBaseBranch(getLocalStorageItem('baseBranch') || '');
    setCurrentBranch(getLocalStorageItem('currentBranch') || '');
    setBaseCommit(getLocalStorageItem('baseCommit') || '');
    setCurrentCommit(getLocalStorageItem('currentCommit') || '');
    setActiveComparison(getLocalStorageItem('activeComparison') as 'branch' | 'commit' | null);
    
    // Load saved selected services
    const savedSelectedServices = getLocalStorageItem('selectedServices');
    if (savedSelectedServices) {
      try {
        const parsedServices = JSON.parse(savedSelectedServices);
        if (Array.isArray(parsedServices)) {
          setSelectedServices(parsedServices);
        }
      } catch (e) {
        console.error('Error parsing saved selected services:', e);
        removeLocalStorageItem('selectedServices');
      }
    }
  }, []);

  // Save state to localStorage when it changes
  useEffect(() => {
    if (selectedRepo) {
      setLocalStorageItem('selectedRepo', JSON.stringify(selectedRepo));
    } else {
      removeLocalStorageItem('selectedRepo');
    }
  }, [selectedRepo]);

  useEffect(() => {
    if (baseBranch) {
      setLocalStorageItem('baseBranch', baseBranch);
    } else {
      removeLocalStorageItem('baseBranch');
    }
  }, [baseBranch]);

  useEffect(() => {
    if (currentBranch) {
      setLocalStorageItem('currentBranch', currentBranch);
    } else {
      removeLocalStorageItem('currentBranch');
    }
  }, [currentBranch]);

  useEffect(() => {
    if (baseCommit) {
      setLocalStorageItem('baseCommit', baseCommit);
    } else {
      removeLocalStorageItem('baseCommit');
    }
  }, [baseCommit]);

  useEffect(() => {
    if (currentCommit) {
      setLocalStorageItem('currentCommit', currentCommit);
    } else {
      removeLocalStorageItem('currentCommit');
    }
  }, [currentCommit]);

  useEffect(() => {
    if (activeComparison) {
      setLocalStorageItem('activeComparison', activeComparison);
    } else {
      removeLocalStorageItem('activeComparison');
    }
  }, [activeComparison]);

  // Save selected services to localStorage when they change
  useEffect(() => {
    if (selectedServices.length > 0) {
      setLocalStorageItem('selectedServices', JSON.stringify(selectedServices));
    } else {
      removeLocalStorageItem('selectedServices');
    }
  }, [selectedServices]);

  // Save state when component unmounts
  useEffect(() => {
    return () => {
      if (selectedServices.length > 0) {
        setLocalStorageItem('selectedServices', JSON.stringify(selectedServices));
      }
    };
  }, [selectedServices]);

  // Load branches when repository is selected (either from new selection or from localStorage)
  useEffect(() => {
    if (selectedRepo) {
      fetchBranches(selectedRepo.id);
    }
  }, [selectedRepo]);

  // Sort and separate branches when allBranches changes
  useEffect(() => {
    if (allBranches.length > 0) {
      // Separate local and remote branches
      const localBranches = allBranches.filter(branch => !branch.isRemote);
      const remoteBranches = allBranches.filter(branch => branch.isRemote);

      // Sort local branches
      const sortedLocalBranches = [...localBranches].sort((a, b) => {
        // Current branch always comes first
        if (a.isCurrent) return -1;
        if (b.isCurrent) return 1;

        // Main and master branches come next
        const isMainOrMaster = (name: string) => name.toLowerCase() === 'main' || name.toLowerCase() === 'master';
        if (isMainOrMaster(a.name) && !isMainOrMaster(b.name)) return -1;
        if (!isMainOrMaster(a.name) && isMainOrMaster(b.name)) return 1;

        // Sort the rest alphabetically
        return a.name.localeCompare(b.name);
      });

      // Sort remote branches alphabetically
      const sortedRemoteBranches = [...remoteBranches].sort((a, b) => a.name.localeCompare(b.name));

      // Set displayed branches
      setDisplayedLocalBranches(sortedLocalBranches.slice(0, BRANCHES_PER_PAGE));
      setDisplayedRemoteBranches(sortedRemoteBranches.slice(0, BRANCHES_PER_PAGE));
    }
  }, [allBranches]);

  // Load base commits when base branch is selected
  useEffect(() => {
    if (selectedRepo && baseBranchForCommits) {
      fetchBaseCommits(selectedRepo.id, baseBranchForCommits);
    } else {
      setBaseCommits([]);
      setDisplayedBaseCommits([]);
      setBaseCommitsPage(1);
    }
  }, [selectedRepo, baseBranchForCommits]);

  // Load current commits when current branch is selected
  useEffect(() => {
    if (selectedRepo && currentBranchForCommits) {
      fetchCurrentCommits(selectedRepo.id, currentBranchForCommits);
    } else {
      setCurrentCommits([]);
      setDisplayedCurrentCommits([]);
      setCurrentCommitsPage(1);
    }
  }, [selectedRepo, currentBranchForCommits]);

  // Update displayed base commits when base commits change or page changes
  useEffect(() => {
    if (baseCommits.length > 0) {
      const endIndex = baseCommitsPage * COMMITS_PER_PAGE;
      setDisplayedBaseCommits(baseCommits.slice(0, endIndex));
    } else {
      setDisplayedBaseCommits([]);
    }
  }, [baseCommits, baseCommitsPage]);

  // Update displayed current commits when current commits change or page changes
  useEffect(() => {
    if (currentCommits.length > 0) {
      const endIndex = currentCommitsPage * COMMITS_PER_PAGE;
      setDisplayedCurrentCommits(currentCommits.slice(0, endIndex));
    } else {
      setDisplayedCurrentCommits([]);
    }
  }, [currentCommits, currentCommitsPage]);

  const handleBaseCommitsScroll = (e: React.UIEvent<HTMLDivElement>) => {
    const target = e.target as HTMLDivElement;
    const { scrollTop, scrollHeight, clientHeight } = target;
    
    // If we're near the bottom (within 50px) and there are more commits to load
    if (scrollHeight - scrollTop - clientHeight < 50 && 
        displayedBaseCommits.length < baseCommits.length) {
      setBaseCommitsPage(prev => prev + 1);
    }
  };

  const handleCurrentCommitsScroll = (e: React.UIEvent<HTMLDivElement>) => {
    const target = e.target as HTMLDivElement;
    const { scrollTop, scrollHeight, clientHeight } = target;
    
    // If we're near the bottom (within 50px) and there are more commits to load
    if (scrollHeight - scrollTop - clientHeight < 50 && 
        displayedCurrentCommits.length < currentCommits.length) {
      setCurrentCommitsPage(prev => prev + 1);
    }
  };

  const fetchBaseCommits = async (repoId: number, branchName: string) => {
    setLoadingBaseCommits(true);
    try {
      const data = await getBranchCommits(repoId, branchName);
      setBaseCommits(data);
    } catch (error) {
      console.error('Error fetching base commits:', error);
      toast({
        title: 'Error',
        description: error instanceof Error ? error.message : 'Failed to fetch base commits',
        variant: 'destructive',
      });
    } finally {
      setLoadingBaseCommits(false);
    }
  };

  const fetchCurrentCommits = async (repoId: number, branchName: string) => {
    setLoadingCurrentCommits(true);
    try {
      const data = await getBranchCommits(repoId, branchName);
      setCurrentCommits(data);
    } catch (error) {
      console.error('Error fetching current commits:', error);
      toast({
        title: 'Error',
        description: error instanceof Error ? error.message : 'Failed to fetch current commits',
        variant: 'destructive',
      });
    } finally {
      setLoadingCurrentCommits(false);
    }
  };

  // Format commit hash to be shorter and more readable
  const formatCommitHash = (hash: string) => {
    return hash.substring(0, 7);
  };

  // Add back the initial repository loading
  useEffect(() => {
    fetchRepositories();
  }, []);

  const fetchRepositories = async () => {
    try {
      const data = await listRepositories();
      setRepositories(data);
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to fetch repositories',
        variant: 'destructive',
      });
    }
  };

  const fetchBranches = async (repoId: number) => {
    setLoadingBranches(true);
    try {
      console.log('Fetching branches for repo:', repoId);
      const data = await getRepositoryBranches(repoId);
      console.log('Fetched branches:', data);
      if (Array.isArray(data)) {
        setAllBranches(data);
      } else {
        console.error('Invalid branch data:', data);
        toast({
          title: 'Error',
          description: 'Invalid branch data received',
          variant: 'destructive',
        });
      }
    } catch (error) {
      console.error('Error fetching branches:', error);
      toast({
        title: 'Error',
        description: error instanceof Error ? error.message : 'Failed to fetch branches',
        variant: 'destructive',
      });
    } finally {
      setLoadingBranches(false);
    }
  };

  const handleRepositoryChange = async (repoId: string) => {
    const repo = repositories.find(r => r.id.toString() === repoId);
    setSelectedRepo(repo || null);
    setBaseBranch('');
    setCurrentBranch('');
    setChangedServices(null);
    setSelectedServices([]); // Clear selected services
    setAllBranches([]);
    setDisplayedLocalBranches([]);
    setDisplayedRemoteBranches([]);
    
    // Clear localStorage for branch-related items when changing repository
    removeLocalStorageItem('baseBranch');
    removeLocalStorageItem('currentBranch');
    removeLocalStorageItem('selectedServices'); // Clear saved selected services
    
    if (repo) {
      console.log('Selected repository:', repo);
      await fetchBranches(repo.id);
    }
  };

  const handleShowMoreLocal = () => {
    const localBranches = allBranches.filter(branch => !branch.isRemote);
    setDisplayedLocalBranches(prev => {
      const currentLength = prev.length;
      const nextBranches = localBranches.slice(currentLength, currentLength + BRANCHES_PER_PAGE);
      return [...prev, ...nextBranches];
    });
  };

  const handleShowMoreRemote = () => {
    const remoteBranches = allBranches.filter(branch => branch.isRemote);
    setDisplayedRemoteBranches(prev => {
      const currentLength = prev.length;
      const nextBranches = remoteBranches.slice(currentLength, currentLength + BRANCHES_PER_PAGE);
      return [...prev, ...nextBranches];
    });
  };

  const clearBranchComparison = () => {
    setBaseBranch('');
    setCurrentBranch('');
    setChangedServices(null);
    setActiveComparison(null);
    setBranchValidationAttempted(false);
    setSelectedServices([]); // Clear selected services
    removeLocalStorageItem('baseBranch');
    removeLocalStorageItem('currentBranch');
    removeLocalStorageItem('activeComparison');
    removeLocalStorageItem('selectedServices'); // Clear saved selected services
  };

  const clearCommitComparison = () => {
    setBaseCommit('');
    setCurrentCommit('');
    setChangedServices(null);
    setActiveComparison(null);
    setCommitValidationAttempted(false);
    setSelectedServices([]); // Clear selected services
    setBaseBranchForCommits('');
    setCurrentBranchForCommits('');
    setBaseCommits([]);
    setCurrentCommits([]);
    setDisplayedBaseCommits([]);
    setDisplayedCurrentCommits([]);
    setBaseCommitsPage(1);
    setCurrentCommitsPage(1);
    removeLocalStorageItem('baseCommit');
    removeLocalStorageItem('currentCommit');
    removeLocalStorageItem('activeComparison');
    removeLocalStorageItem('selectedServices'); // Clear saved selected services
  };

  const handleDetectChanges = async () => {
    setBranchValidationAttempted(true);
    
    if (!selectedRepo || !baseBranch || !currentBranch) {
      toast({
        title: 'Missing Selection',
        description: 'Please select a repository and both branches to compare',
        variant: 'destructive',
      });
      return;
    }

    if (baseBranch === currentBranch) {
      toast({
        title: 'Invalid Selection',
        description: 'Please select different branches to compare',
        variant: 'destructive',
      });
      return;
    }

    setLoading(true);
    setChangedServices(null);
    setActiveComparison('branch');
    
    try {
      const result = await detectChanges({
        url: selectedRepo.url,
        baseBranch,
        currentBranch,
      });
      
      setChangedServices(result.changedServices);
      
      if (!result.changedServices || result.changedServices.length === 0) {
        toast({
          title: 'No Changes Detected',
          description: `No service changes found between '${baseBranch}' and '${currentBranch}'`,
          variant: 'default',
        });
      } else {
        toast({
          title: 'Changes Detected',
          description: `Found changes in ${result.changedServices.length} service${result.changedServices.length === 1 ? '' : 's'}`,
          variant: 'default',
        });
      }
    } catch (error) {
      console.error('Error detecting changes:', error);
      toast({
        title: 'Error Detecting Changes',
        description: error instanceof Error ? error.message : 'Failed to detect changes between branches',
        variant: 'destructive',
      });
      setChangedServices(null);
    } finally {
      setLoading(false);
    }
  };

  const handleDetectCommitChanges = async () => {
    setCommitValidationAttempted(true);
    
    if (!selectedRepo || !baseCommit || !currentCommit) {
      toast({
        title: 'Missing Selection',
        description: 'Please select a repository and provide both commit hashes',
        variant: 'destructive',
      });
      return;
    }

    if (baseCommit === currentCommit) {
      toast({
        title: 'Invalid Selection',
        description: 'Please provide different commit hashes to compare',
        variant: 'destructive',
      });
      return;
    }

    setLoading(true);
    setChangedServices(null);
    setActiveComparison('commit');
    
    try {
      const result = await detectCommitChanges({
        url: selectedRepo.url,
        baseCommit,
        currentCommit,
      });
      
      setChangedServices(result.changedServices);
      
      if (!result.changedServices || result.changedServices.length === 0) {
        toast({
          title: 'No Changes Detected',
          description: `No service changes found between commits`,
          variant: 'default',
        });
      } else {
        toast({
          title: 'Changes Detected',
          description: `Found changes in ${result.changedServices.length} service${result.changedServices.length === 1 ? '' : 's'}`,
          variant: 'default',
        });
      }
    } catch (error) {
      console.error('Error detecting commit changes:', error);
      toast({
        title: 'Error Detecting Changes',
        description: error instanceof Error ? error.message : 'Failed to detect changes between commits',
        variant: 'destructive',
      });
      setChangedServices(null);
    } finally {
      setLoading(false);
    }
  };

  const getFilteredBranches = (branches: Branch[], excludeBranch?: string) => {
    return branches.filter(branch => branch.name !== excludeBranch);
  };

  // Handle service selection with persistence
  const handleServiceSelection = (servicePath: string) => {
    setSelectedServices(prev => {
      const newSelection = prev.includes(servicePath)
        ? prev.filter(s => s !== servicePath)
        : [...prev, servicePath];
      
      // Save to localStorage immediately
      if (newSelection.length > 0) {
        setLocalStorageItem('selectedServices', JSON.stringify(newSelection));
      } else {
        removeLocalStorageItem('selectedServices');
      }
      
      return newSelection;
    });
  };

  // Handle select all with persistence
  const handleSelectAllServices = () => {
    if (changedServices && changedServices.length > 0) {
      const availableServices = changedServices
        .filter(service => service.hasDockerfile)
        .map(service => service.path);
      
      if (selectedServices.length === availableServices.length) {
        // If all are selected, deselect all
        setSelectedServices([]);
        removeLocalStorageItem('selectedServices');
      } else {
        // Otherwise, select all available services
        setSelectedServices([...availableServices]);
        setLocalStorageItem('selectedServices', JSON.stringify(availableServices));
      }
    }
  };

  const handleBuildSelectedServices = async () => {
    if (!selectedRepo || selectedServices.length === 0) return;

    setIsBuilding(true);
    
    try {
      const result = await buildImage({
        url: selectedRepo.url,
        branch: currentBranch || currentCommit,
        servicePaths: selectedServices,
        tag: buildTag,
        registry: registry,
      });

      // Check if the response indicates success
      if (result && result.success) {
        toast({
          title: 'Build Successful',
          description: `Successfully built images for ${selectedServices.length} service${selectedServices.length === 1 ? '' : 's'}`,
          variant: 'default',
        });
      } else {
        // Handle case where API returned success but with error message
        toast({
          title: 'Build Failed',
          description: result?.message || 'Failed to start build',
          variant: 'destructive',
        });
      }
    } catch (error) {
      console.error('Error building images:', error);
      toast({
        title: 'Build Error',
        description: error instanceof Error ? error.message : 'Failed to build images',
        variant: 'destructive',
      });
    } finally {
      setIsBuilding(false);
    }
  };

  return (
    <div className="container mx-auto py-10 space-y-8">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Image Builder</h1>
      </div>

      {/* Repository Selection */}
      <div className="bg-card rounded-lg border overflow-hidden">
        <div className="p-6">
          <div className="space-y-4">
            <h2 className="text-2xl font-bold tracking-tight">Repository Selection</h2>
            {isClient && (
              <Select
                value={selectedRepo?.id.toString()}
                onValueChange={handleRepositoryChange}
              >
                <SelectTrigger className="w-full h-12">
                  <SelectValue placeholder="Select a repository" />
                </SelectTrigger>
                <SelectContent>
                  {repositories.map((repo) => (
                    <SelectItem key={repo.id} value={repo.id.toString()}>
                      <div className="flex items-center gap-2">
                        <GitBranch className="h-4 w-4 text-muted-foreground" />
                        <span>{repo.name}</span>
                      </div>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
            {repositories.length === 0 && (
              <p className="text-sm text-muted-foreground">
                No repositories found. Please add repositories in the Repositories tab.
              </p>
            )}
          </div>
        </div>
      </div>

      {/* Branch Selection */}
      <div className="bg-card rounded-lg border overflow-hidden">
        <div className="p-6 border-b">
          <div className="flex items-center justify-between">
            <h2 className="text-2xl font-bold tracking-tight">Branch Comparison</h2>
            <div className="flex items-center gap-2">
              <Button
                onClick={clearBranchComparison}
                variant="outline"
                size="sm"
                className="h-8"
                disabled={loading || activeComparison === 'commit'}
              >
                <X className="h-4 w-4 mr-1" />
                Clear
              </Button>
              <Button
                onClick={handleDetectChanges}
                disabled={!selectedRepo || !baseBranch || !currentBranch || loading || baseBranch === currentBranch || activeComparison === 'commit'}
                size="lg"
                variant={loading ? "outline" : "default"}
                className="min-w-[200px]"
              >
                <div className="flex items-center justify-center gap-2">
                  {loading ? (
                    <>
                      <Loader2 className="h-4 w-4 animate-spin" />
                      <span>Detecting Changes...</span>
                    </>
                  ) : (
                    <>
                      <GitCompare className="h-4 w-4" />
                      <span>Compare Branches</span>
                      <ArrowRight className="h-4 w-4 ml-1" />
                    </>
                  )}
                </div>
              </Button>
            </div>
          </div>
        </div>

        <div className="p-6 grid gap-8">
          <div className="grid grid-cols-2 gap-8">
            <div className="space-y-4">
              <div className="flex items-center gap-2">
                <GitBranch className="h-5 w-5 text-muted-foreground" />
                <h3 className="text-lg font-medium">Base Branch</h3>
              </div>
              {isClient && (
                <BranchSelect
                  value={baseBranch}
                  onValueChange={setBaseBranch}
                  placeholder="Select base branch"
                  excludeBranch={currentBranch}
                  displayedLocalBranches={displayedLocalBranches}
                  displayedRemoteBranches={displayedRemoteBranches}
                  allBranches={allBranches}
                  loadingBranches={loadingBranches}
                  onShowMoreLocal={handleShowMoreLocal}
                  onShowMoreRemote={handleShowMoreRemote}
                  disabled={activeComparison === 'commit'}
                />
              )}
            </div>

            <div className="space-y-4">
              <div className="flex items-center gap-2">
                <GitBranch className="h-5 w-5 text-muted-foreground" />
                <h3 className="text-lg font-medium">Current Branch</h3>
              </div>
              {isClient && (
                <BranchSelect
                  value={currentBranch}
                  onValueChange={setCurrentBranch}
                  placeholder="Select current branch"
                  excludeBranch={baseBranch}
                  displayedLocalBranches={displayedLocalBranches}
                  displayedRemoteBranches={displayedRemoteBranches}
                  allBranches={allBranches}
                  loadingBranches={loadingBranches}
                  onShowMoreLocal={handleShowMoreLocal}
                  onShowMoreRemote={handleShowMoreRemote}
                  disabled={activeComparison === 'commit'}
                />
              )}
            </div>
          </div>

          <div className="flex justify-center">
            {branchValidationAttempted && (
              <>
                {(!baseBranch || !currentBranch) && (
                  <p className="text-sm text-muted-foreground">
                    Select both branches to compare changes
                  </p>
                )}
                {baseBranch === currentBranch && baseBranch !== '' && (
                  <p className="text-sm text-destructive">
                    Please select different branches to compare
                  </p>
                )}
              </>
            )}
            {activeComparison === 'commit' && (
              <p className="text-sm text-muted-foreground">
                Clear commit comparison to use branch comparison
              </p>
            )}
          </div>
        </div>
      </div>

      {/* Commit Selection */}
      <div className="bg-card rounded-lg border overflow-hidden">
        <div className="p-6 border-b">
          <div className="flex items-center justify-between">
            <h2 className="text-2xl font-bold tracking-tight">Commit Comparison</h2>
            <div className="flex items-center gap-2">
              <Button
                onClick={clearCommitComparison}
                variant="outline"
                size="sm"
                className="h-8"
                disabled={loading || activeComparison === 'branch'}
              >
                <X className="h-4 w-4 mr-1" />
                Clear
              </Button>
              <Button
                onClick={handleDetectCommitChanges}
                disabled={!selectedRepo || !baseCommit || !currentCommit || loading || baseCommit === currentCommit || activeComparison === 'branch'}
                size="lg"
                variant={loading ? "outline" : "default"}
                className="min-w-[200px]"
              >
                <div className="flex items-center justify-center gap-2">
                  {loading ? (
                    <>
                      <Loader2 className="h-4 w-4 animate-spin" />
                      <span>Detecting Changes...</span>
                    </>
                  ) : (
                    <>
                      <GitCompare className="h-4 w-4" />
                      <span>Compare Commits</span>
                      <ArrowRight className="h-4 w-4 ml-1" />
                    </>
                  )}
                </div>
              </Button>
            </div>
          </div>
        </div>

        <div className="p-6 grid gap-8">
          <div className="grid grid-cols-2 gap-8">
            {/* Base Branch and Commit Selection */}
            <div className="space-y-4">
              <div className="flex items-center gap-2">
                <GitBranch className="h-5 w-5 text-muted-foreground" />
                <h3 className="text-lg font-medium">Base Branch</h3>
              </div>
              {isClient && (
                <BranchSelect
                  value={baseBranchForCommits}
                  onValueChange={setBaseBranchForCommits}
                  placeholder="Select base branch"
                  excludeBranch=""
                  displayedLocalBranches={displayedLocalBranches}
                  displayedRemoteBranches={displayedRemoteBranches}
                  allBranches={allBranches}
                  loadingBranches={loadingBranches}
                  onShowMoreLocal={handleShowMoreLocal}
                  onShowMoreRemote={handleShowMoreRemote}
                  disabled={activeComparison === 'branch'}
                />
              )}

              {baseBranchForCommits && (
                <div className="space-y-4 mt-4">
                  <div className="flex items-center gap-2">
                    <GitCommit className="h-5 w-5 text-muted-foreground" />
                    <h3 className="text-lg font-medium">Base Commit</h3>
                  </div>
                  {isClient && (
                    <Select
                      value={baseCommit}
                      onValueChange={setBaseCommit}
                      disabled={activeComparison === 'branch' || loadingBaseCommits}
                    >
                      <SelectTrigger className="h-12">
                        <SelectValue placeholder={loadingBaseCommits ? "Loading commits..." : "Select base commit"} />
                      </SelectTrigger>
                      <SelectContent>
                        <div 
                          className="max-h-[300px] overflow-y-auto" 
                          onScroll={handleBaseCommitsScroll}
                        >
                          {displayedBaseCommits.map((commit) => (
                            <SelectItem key={commit.hash} value={commit.hash}>
                              <div className="flex flex-col">
                                <span className="font-medium truncate">{commit.message}</span>
                                <span className="text-xs text-muted-foreground">
                                  {formatCommitHash(commit.hash)} • {new Date(commit.timestamp).toLocaleString()}
                                </span>
                              </div>
                            </SelectItem>
                          ))}
                          {displayedBaseCommits.length < baseCommits.length && (
                            <div className="p-2 text-center text-sm text-muted-foreground">
                              Scroll down to load more commits...
                            </div>
                          )}
                        </div>
                      </SelectContent>
                    </Select>
                  )}
                </div>
              )}
            </div>

            {/* Current Branch and Commit Selection */}
            <div className="space-y-4">
              <div className="flex items-center gap-2">
                <GitBranch className="h-5 w-5 text-muted-foreground" />
                <h3 className="text-lg font-medium">Current Branch</h3>
              </div>
              {isClient && (
                <BranchSelect
                  value={currentBranchForCommits}
                  onValueChange={setCurrentBranchForCommits}
                  placeholder="Select current branch"
                  excludeBranch=""
                  displayedLocalBranches={displayedLocalBranches}
                  displayedRemoteBranches={displayedRemoteBranches}
                  allBranches={allBranches}
                  loadingBranches={loadingBranches}
                  onShowMoreLocal={handleShowMoreLocal}
                  onShowMoreRemote={handleShowMoreRemote}
                  disabled={activeComparison === 'branch'}
                />
              )}

              {currentBranchForCommits && (
                <div className="space-y-4 mt-4">
                  <div className="flex items-center gap-2">
                    <GitCommit className="h-5 w-5 text-muted-foreground" />
                    <h3 className="text-lg font-medium">Current Commit</h3>
                  </div>
                  {isClient && (
                    <Select
                      value={currentCommit}
                      onValueChange={setCurrentCommit}
                      disabled={activeComparison === 'branch' || loadingCurrentCommits}
                    >
                      <SelectTrigger className="h-12">
                        <SelectValue placeholder={loadingCurrentCommits ? "Loading commits..." : "Select current commit"} />
                      </SelectTrigger>
                      <SelectContent>
                        <div 
                          className="max-h-[300px] overflow-y-auto" 
                          onScroll={handleCurrentCommitsScroll}
                        >
                          {displayedCurrentCommits.map((commit) => (
                            <SelectItem key={commit.hash} value={commit.hash}>
                              <div className="flex flex-col">
                                <span className="font-medium truncate">{commit.message}</span>
                                <span className="text-xs text-muted-foreground">
                                  {formatCommitHash(commit.hash)} • {new Date(commit.timestamp).toLocaleString()}
                                </span>
                              </div>
                            </SelectItem>
                          ))}
                          {displayedCurrentCommits.length < currentCommits.length && (
                            <div className="p-2 text-center text-sm text-muted-foreground">
                              Scroll down to load more commits...
                            </div>
                          )}
                        </div>
                      </SelectContent>
                    </Select>
                  )}
                </div>
              )}
            </div>
          </div>

          <div className="flex justify-center">
            {commitValidationAttempted && (
              <>
                {(!baseCommit || !currentCommit) && (
                  <p className="text-sm text-muted-foreground">
                    Select both commits to compare changes
                  </p>
                )}
                {baseCommit === currentCommit && baseCommit !== '' && (
                  <p className="text-sm text-destructive">
                    Please select different commits to compare
                  </p>
                )}
              </>
            )}
            {activeComparison === 'branch' && (
              <p className="text-sm text-muted-foreground">
                Clear branch comparison to use commit comparison
              </p>
            )}
            {(!baseBranchForCommits || !currentBranchForCommits) && (
              <p className="text-sm text-muted-foreground">
                Select both branches to view and select commits
              </p>
            )}
          </div>
        </div>
      </div>

      {/* Changed Services */}
      {changedServices !== null && (
        <div className="grid gap-4 p-6 bg-card rounded-lg border">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold">Changed Services</h2>
            {changedServices.length > 0 && (
              <div className="text-sm text-muted-foreground">
                {changedServices.length} service{changedServices.length === 1 ? '' : 's'} changed
              </div>
            )}
          </div>

          {/* Build Configuration */}
          {changedServices && changedServices.length > 0 && (
            <div className="grid grid-cols-2 gap-4 p-4 bg-muted rounded-lg">
              <div className="space-y-2">
                <label className="text-sm font-medium">Build Tag</label>
                <Input
                  value={buildTag}
                  onChange={(e) => setBuildTag(e.target.value)}
                  placeholder="e.g. v1.0.0"
                  className="h-9"
                />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">Registry</label>
                <Input
                  value={registry}
                  onChange={(e) => setRegistry(e.target.value)}
                  placeholder="e.g. localhost:5000"
                  className="h-9"
                />
              </div>
            </div>
          )}

          {changedServices && changedServices.length > 0 ? (
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <Button 
                    variant="outline" 
                    size="sm" 
                    onClick={handleSelectAllServices}
                    className="h-8"
                  >
                    {selectedServices.length === changedServices.filter(s => s.hasDockerfile).length ? (
                      <>
                        <CheckSquare className="h-4 w-4 mr-2" />
                        Deselect All
                      </>
                    ) : (
                      <>
                        <Square className="h-4 w-4 mr-2" />
                        Select All
                      </>
                    )}
                  </Button>
                  <span className="text-sm text-muted-foreground">
                    {selectedServices.length} of {changedServices.filter(s => s.hasDockerfile).length} selected
                  </span>
                </div>
                <Button
                  onClick={handleBuildSelectedServices}
                  disabled={selectedServices.length === 0 || isBuilding}
                  size="sm"
                >
                  {isBuilding ? (
                    <>
                      <Loader2 className="h-4 w-4 animate-spin mr-2" />
                      Building...
                    </>
                  ) : (
                    <>
                      <Package className="h-4 w-4 mr-2" />
                      Build Selected ({selectedServices.length})
                    </>
                  )}
                </Button>
              </div>
              
              <div className="grid gap-2">
                {changedServices.map((service) => (
                  <div
                    key={service.path}
                    className="flex items-center justify-between gap-2 p-3 bg-muted rounded-lg hover:bg-muted/80 transition-colors"
                  >
                    <div className="flex items-center gap-2">
                      <Checkbox 
                        id={`service-${service.path}`}
                        checked={selectedServices.includes(service.path)}
                        onCheckedChange={() => handleServiceSelection(service.path)}
                        disabled={!service.hasDockerfile}
                      />
                      <label 
                        htmlFor={`service-${service.path}`}
                        className={cn(
                          "flex items-center gap-2",
                          !service.hasDockerfile && "cursor-not-allowed opacity-50"
                        )}
                      >
                        <GitBranch className="h-4 w-4 text-muted-foreground" />
                        <span className="font-medium">{service.path}</span>
                        {!service.hasDockerfile && (
                          <span className="text-sm text-destructive flex items-center gap-1">
                            <AlertCircle className="h-4 w-4" />
                            No Dockerfile found
                          </span>
                        )}
                      </label>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          ) : (
            <div className="flex items-center gap-2 p-4 bg-muted rounded-lg text-muted-foreground">
              <AlertCircle className="h-5 w-5" />
              <span>No service changes detected between <code className="text-sm">{baseBranch || baseCommit}</code> and <code className="text-sm">{currentBranch || currentCommit}</code></span>
            </div>
          )}
        </div>
      )}
    </div>
  );
} 