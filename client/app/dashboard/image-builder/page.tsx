'use client';

import { useState, useEffect } from 'react';
import { Repository, Branch } from '@/lib/types';
import { listRepositories, getRepositoryBranches, detectChanges, detectCommitChanges } from '@/lib/api';
import { Button } from '@/components/ui/button';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useToast } from '@/hooks/use-toast';
import { GitBranch, ArrowRight, ChevronDown, Globe, Laptop, AlertCircle, Loader2, GitCompare, GitCommit, X } from 'lucide-react';
import { Input } from "@/components/ui/input";

interface ChangedService {
  name: string;
  path: string;
}

const BRANCHES_PER_PAGE = 20;

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
  const [selectedRepo, setSelectedRepo] = useState<Repository | null>(() => {
    const saved = localStorage.getItem('selectedRepo');
    return saved ? JSON.parse(saved) : null;
  });
  const [allBranches, setAllBranches] = useState<Branch[]>([]);
  const [displayedLocalBranches, setDisplayedLocalBranches] = useState<Branch[]>([]);
  const [displayedRemoteBranches, setDisplayedRemoteBranches] = useState<Branch[]>([]);
  const [baseBranch, setBaseBranch] = useState<string>(() => {
    return localStorage.getItem('baseBranch') || '';
  });
  const [currentBranch, setCurrentBranch] = useState<string>(() => {
    return localStorage.getItem('currentBranch') || '';
  });
  const [baseCommit, setBaseCommit] = useState<string>(() => {
    return localStorage.getItem('baseCommit') || '';
  });
  const [currentCommit, setCurrentCommit] = useState<string>(() => {
    return localStorage.getItem('currentCommit') || '';
  });
  const [changedServices, setChangedServices] = useState<string[] | null>(null);
  const [loading, setLoading] = useState(false);
  const [loadingBranches, setLoadingBranches] = useState(false);
  const [activeComparison, setActiveComparison] = useState<'branch' | 'commit' | null>(() => {
    return localStorage.getItem('activeComparison') as 'branch' | 'commit' | null;
  });
  const [branchValidationAttempted, setBranchValidationAttempted] = useState(false);
  const [commitValidationAttempted, setCommitValidationAttempted] = useState(false);
  const { toast } = useToast();

  // Save state to localStorage when it changes
  useEffect(() => {
    if (selectedRepo) {
      localStorage.setItem('selectedRepo', JSON.stringify(selectedRepo));
    } else {
      localStorage.removeItem('selectedRepo');
    }
  }, [selectedRepo]);

  useEffect(() => {
    if (baseBranch) {
      localStorage.setItem('baseBranch', baseBranch);
    } else {
      localStorage.removeItem('baseBranch');
    }
  }, [baseBranch]);

  useEffect(() => {
    if (currentBranch) {
      localStorage.setItem('currentBranch', currentBranch);
    } else {
      localStorage.removeItem('currentBranch');
    }
  }, [currentBranch]);

  useEffect(() => {
    if (baseCommit) {
      localStorage.setItem('baseCommit', baseCommit);
    } else {
      localStorage.removeItem('baseCommit');
    }
  }, [baseCommit]);

  useEffect(() => {
    if (currentCommit) {
      localStorage.setItem('currentCommit', currentCommit);
    } else {
      localStorage.removeItem('currentCommit');
    }
  }, [currentCommit]);

  useEffect(() => {
    if (activeComparison) {
      localStorage.setItem('activeComparison', activeComparison);
    } else {
      localStorage.removeItem('activeComparison');
    }
  }, [activeComparison]);

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
    setAllBranches([]);
    setDisplayedLocalBranches([]);
    setDisplayedRemoteBranches([]);
    
    // Clear localStorage for branch-related items when changing repository
    localStorage.removeItem('baseBranch');
    localStorage.removeItem('currentBranch');
    
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
    localStorage.removeItem('baseBranch');
    localStorage.removeItem('currentBranch');
    localStorage.removeItem('activeComparison');
  };

  const clearCommitComparison = () => {
    setBaseCommit('');
    setCurrentCommit('');
    setChangedServices(null);
    setActiveComparison(null);
    setCommitValidationAttempted(false);
    localStorage.removeItem('baseCommit');
    localStorage.removeItem('currentCommit');
    localStorage.removeItem('activeComparison');
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
            </div>

            <div className="space-y-4">
              <div className="flex items-center gap-2">
                <GitBranch className="h-5 w-5 text-muted-foreground" />
                <h3 className="text-lg font-medium">Current Branch</h3>
              </div>
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
            <div className="space-y-4">
              <div className="flex items-center gap-2">
                <GitCommit className="h-5 w-5 text-muted-foreground" />
                <h3 className="text-lg font-medium">Base Commit</h3>
              </div>
              <Input
                placeholder="Enter base commit hash"
                value={baseCommit}
                onChange={(e) => setBaseCommit(e.target.value)}
                className="h-12"
                disabled={activeComparison === 'branch'}
              />
            </div>

            <div className="space-y-4">
              <div className="flex items-center gap-2">
                <GitCommit className="h-5 w-5 text-muted-foreground" />
                <h3 className="text-lg font-medium">Current Commit</h3>
              </div>
              <Input
                placeholder="Enter current commit hash"
                value={currentCommit}
                onChange={(e) => setCurrentCommit(e.target.value)}
                className="h-12"
                disabled={activeComparison === 'branch'}
              />
            </div>
          </div>

          <div className="flex justify-center">
            {commitValidationAttempted && (
              <>
                {(!baseCommit || !currentCommit) && (
                  <p className="text-sm text-muted-foreground">
                    Enter both commit hashes to compare changes
                  </p>
                )}
                {baseCommit === currentCommit && baseCommit !== '' && (
                  <p className="text-sm text-destructive">
                    Please provide different commit hashes to compare
                  </p>
                )}
              </>
            )}
            {activeComparison === 'branch' && (
              <p className="text-sm text-muted-foreground">
                Clear branch comparison to use commit comparison
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
          {changedServices && changedServices.length > 0 ? (
            <div className="grid gap-2">
              {changedServices.map((service) => (
                <div
                  key={service}
                  className="flex items-center gap-2 p-3 bg-muted rounded-lg hover:bg-muted/80 transition-colors"
                >
                  <GitBranch className="h-4 w-4 text-muted-foreground" />
                  <span className="font-medium">{service}</span>
                </div>
              ))}
            </div>
          ) : (
            <div className="flex items-center gap-2 p-4 bg-muted rounded-lg text-muted-foreground">
              <AlertCircle className="h-5 w-5" />
              <span>No service changes detected between <code className="text-sm">{baseBranch}</code> and <code className="text-sm">{currentBranch}</code></span>
            </div>
          )}
        </div>
      )}
    </div>
  );
} 