'use client';

import { useEffect, useState } from 'react';
import { Repository, Branch, Commit } from '@/lib/types';
import { listRepositories, deleteRepository, getRepositoryBranches, getBranchCommits, checkoutBranch, syncRepository } from '@/lib/api';
import { Button } from '@/components/ui/button';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { useToast } from '@/hooks/use-toast';
import { format, parseISO } from 'date-fns';
import { AddRepositoryDialog } from './components/add-repository-dialog';
import { ChevronDown, GitBranch, Globe, Laptop, MoreHorizontal, ChevronUp, GitCommit, ChevronRight, RotateCw } from 'lucide-react';
import { cn } from '@/lib/utils';

const formatDate = (dateString: string) => {
  try {
    const date = new Date(dateString);
    return {
      date: format(date, 'PPp'),
      time: format(date, 'HH:mm:ss')
    };
  } catch (error) {
    console.error('Date parsing error:', error);
    return {
      date: dateString,
      time: ''
    };
  }
};

const CommitDetails = ({ commit }: { commit: Commit }) => {
  const { date, time } = formatDate(commit.timestamp);

  return (
    <div className="pl-8 pr-2 py-2 space-y-3 bg-muted/30 rounded-lg">
      <div className="grid grid-cols-2 gap-4">
        <div>
          <div className="text-sm font-medium text-muted-foreground">Hash</div>
          <div className="font-mono text-sm">{commit.hash}</div>
        </div>
        <div>
          <div className="text-sm font-medium text-muted-foreground">Author</div>
          <div className="text-sm">{commit.author}</div>
        </div>
      </div>
      <div>
        <div className="text-sm font-medium text-muted-foreground">Message</div>
        <div className="text-sm whitespace-pre-wrap">{commit.message}</div>
      </div>
      <div>
        <div className="text-sm font-medium text-muted-foreground">Date</div>
        <div className="text-sm">{date}</div>
        <div className="text-sm text-muted-foreground">{time}</div>
      </div>
    </div>
  );
};

const CommitList = ({ commits }: { commits: Commit[] }) => {
  const [displayCount, setDisplayCount] = useState(5);
  const [expandedCommit, setExpandedCommit] = useState<string | null>(null);
  const displayedCommits = commits.slice(0, displayCount);
  const hasMore = commits.length > displayCount;
  const remainingCount = commits.length - displayCount;

  const handleShowMore = () => {
    setDisplayCount(prev => Math.min(prev + 5, commits.length));
  };

  const handleShowLess = () => {
    setDisplayCount(5);
  };

  return (
    <div className="mt-2 space-y-2">
      {displayedCommits.map((commit) => {
        const { date, time } = formatDate(commit.timestamp);
        const isExpanded = expandedCommit === commit.hash;
        return (
          <div key={commit.hash} className="space-y-1">
            <div
              className="flex items-center justify-between p-2 rounded-lg bg-muted/50 cursor-pointer hover:bg-muted transition-colors"
              onClick={() => setExpandedCommit(isExpanded ? null : commit.hash)}
            >
              <div className="flex items-center gap-2">
                <ChevronRight className={cn(
                  "h-4 w-4 text-muted-foreground transition-transform",
                  isExpanded && "transform rotate-90"
                )} />
                <GitCommit className="h-4 w-4 text-muted-foreground" />
                <div>
                  <div className="font-medium line-clamp-1">{commit.message}</div>
                  <div className="text-sm text-muted-foreground">{commit.author}</div>
                </div>
              </div>
              <div className="text-sm text-muted-foreground text-right">
                <div>{date}</div>
                <div>{time}</div>
              </div>
            </div>
            {isExpanded && <CommitDetails commit={commit} />}
          </div>
        );
      })}
      {hasMore && (
        <div className="flex flex-col gap-2">
          <Button
            variant="ghost"
            size="sm"
            className="w-full flex items-center justify-center gap-2 text-muted-foreground hover:text-foreground"
            onClick={handleShowMore}
          >
            <MoreHorizontal className="h-4 w-4" />
            Show 5 More Commits
          </Button>
          <Button
            variant="ghost"
            size="sm"
            className="w-full flex items-center justify-center gap-2 text-muted-foreground hover:text-foreground"
            onClick={() => setDisplayCount(commits.length)}
          >
            <MoreHorizontal className="h-4 w-4" />
            Show All Commits ({remainingCount} remaining)
          </Button>
        </div>
      )}
      {displayCount > 5 && (
        <Button
          variant="ghost"
          size="sm"
          className="w-full flex items-center justify-center gap-2 text-muted-foreground hover:text-foreground"
          onClick={handleShowLess}
        >
          <ChevronUp className="h-4 w-4" />
          Show Less
        </Button>
      )}
    </div>
  );
};

const BranchList = ({ branches, title, icon: Icon, repositoryId, onBranchesChange }: { 
  branches: Branch[], 
  title: string, 
  icon: React.ElementType, 
  repositoryId: number,
  onBranchesChange: () => void 
}) => {
  const [displayCount, setDisplayCount] = useState(10);
  const [selectedBranch, setSelectedBranch] = useState<string | null>(null);
  const [commits, setCommits] = useState<Commit[]>([]);
  const [loadingCommits, setLoadingCommits] = useState(false);
  const [checkoutLoading, setCheckoutLoading] = useState<string | null>(null);
  const [localBranches, setLocalBranches] = useState<Branch[]>(branches);
  const { toast } = useToast();

  // Update local branches when props change
  useEffect(() => {
    setLocalBranches(branches);
  }, [branches]);

  // Sort branches to show current branch first
  const sortedBranches = [...localBranches].sort((a, b) => {
    if (a.isCurrent) return -1;
    if (b.isCurrent) return 1;
    return 0;
  });

  const displayedBranches = sortedBranches.slice(0, displayCount);
  const hasMore = sortedBranches.length > displayCount;
  const remainingCount = sortedBranches.length - displayCount;

  const fetchRepositories = async () => {
    try {
      const repos = await listRepositories();
      // Update the branches list with the new current branch information
      const updatedBranches = localBranches.map(branch => ({
        ...branch,
        isCurrent: repos.find(repo => repo.id === repositoryId)?.currentBranch === branch.name
      }));
      setLocalBranches(updatedBranches);
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to refresh repository information',
        variant: 'destructive',
      });
    }
  };

  const handleShowMore = () => {
    setDisplayCount(prev => Math.min(prev + 10, localBranches.length));
  };

  const handleShowAll = () => {
    setDisplayCount(localBranches.length);
  };

  const handleShowLess = () => {
    setDisplayCount(10);
  };

  const handleBranchClick = async (branch: Branch) => {
    if (selectedBranch === branch.name) {
      setSelectedBranch(null);
      return;
    }

    setSelectedBranch(branch.name);
    setLoadingCommits(true);
    try {
      const branchCommits = await getBranchCommits(repositoryId, branch.name);
      setCommits(branchCommits);
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to fetch branch commits',
        variant: 'destructive',
      });
    } finally {
      setLoadingCommits(false);
    }
  };

  const handleCheckout = async (branch: Branch) => {
    if (branch.isCurrent) return;
    
    setCheckoutLoading(branch.name);
    try {
      await checkoutBranch(repositoryId, branch.name);
      toast({
        title: 'Success',
        description: `Switched to branch ${branch.name}`,
      });
      
      // Notify parent component to reload branches
      onBranchesChange();
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : String(error);
      toast({
        title: 'Error',
        description: errorMessage,
        variant: 'destructive',
      });
    } finally {
      setCheckoutLoading(null);
    }
  };

  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2 mb-2">
        <Icon className="h-4 w-4" />
        <h4 className="font-medium text-sm">{title}</h4>
      </div>
      <div className="pl-6 space-y-2">
        {displayedBranches.map((branch) => (
          <div key={branch.name}>
            <div
              className={cn(
                "flex items-center justify-between p-2 rounded-lg bg-background cursor-pointer hover:bg-muted/50 transition-colors",
                selectedBranch === branch.name && "bg-muted"
              )}
              onClick={() => handleBranchClick(branch)}
            >
              <div className="flex items-center gap-2">
                <span className="font-medium">{branch.name}</span>
                {branch.isCurrent && (
                  <span className="text-xs px-2 py-1 rounded-full bg-primary/10 text-primary">
                    Current
                  </span>
                )}
              </div>
              <div className="flex items-center gap-2">
                <div className="text-sm text-muted-foreground">
                  <div>{branch.lastCommit}</div>
                  <div>{formatDate(branch.lastCommitDate).date}</div>
                </div>
                {!branch.isCurrent && (
                  <Button
                    variant="ghost"
                    size="sm"
                    className="ml-2"
                    onClick={(e) => {
                      e.stopPropagation();
                      handleCheckout(branch);
                    }}
                    disabled={!!checkoutLoading}
                  >
                    {checkoutLoading === branch.name ? (
                      <div className="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
                    ) : (
                      'Checkout'
                    )}
                  </Button>
                )}
              </div>
            </div>
            {selectedBranch === branch.name && (
              <div className="pl-4">
                {loadingCommits ? (
                  <div className="text-sm text-muted-foreground p-2">Loading commits...</div>
                ) : commits.length > 0 ? (
                  <CommitList commits={commits} />
                ) : (
                  <div className="text-sm text-muted-foreground p-2">No commits found</div>
                )}
              </div>
            )}
          </div>
        ))}
        {hasMore && (
          <div className="flex flex-col gap-2">
            <Button
              variant="ghost"
              size="sm"
              className="w-full flex items-center justify-center gap-2 text-muted-foreground hover:text-foreground"
              onClick={handleShowMore}
            >
              <MoreHorizontal className="h-4 w-4" />
              Show 10 More Branches
            </Button>
            <Button
              variant="ghost"
              size="sm"
              className="w-full flex items-center justify-center gap-2 text-muted-foreground hover:text-foreground"
              onClick={handleShowAll}
            >
              <MoreHorizontal className="h-4 w-4" />
              Show All Branches ({remainingCount} remaining)
            </Button>
          </div>
        )}
        {displayCount > 10 && (
          <Button
            variant="ghost"
            size="sm"
            className="w-full flex items-center justify-center gap-2 text-muted-foreground hover:text-foreground"
            onClick={handleShowLess}
          >
            <ChevronUp className="h-4 w-4" />
            Show Less
          </Button>
        )}
      </div>
    </div>
  );
};

export default function RepositoriesPage() {
  const [repositories, setRepositories] = useState<Repository[]>([]);
  const [loading, setLoading] = useState(true);
  const [expandedRepo, setExpandedRepo] = useState<number | null>(null);
  const [syncingRepo, setSyncingRepo] = useState<number | null>(null);
  const { toast } = useToast();

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
    } finally {
      setLoading(false);
    }
  };

  const fetchBranches = async (repoId: number) => {
    try {
      const branches = await getRepositoryBranches(repoId);
      setRepositories(prev => 
        prev.map(repo => 
          repo.id === repoId 
            ? { ...repo, branches } 
            : repo
        )
      );
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to fetch repository branches',
        variant: 'destructive',
      });
    }
  };

  const handleBranchesChange = async (repoId: number) => {
    await fetchBranches(repoId);
  };

  useEffect(() => {
    fetchRepositories();
  }, []);

  const handleDelete = async (id: number) => {
    try {
      await deleteRepository(id);
      toast({
        title: 'Success',
        description: 'Repository deleted successfully',
      });
      fetchRepositories();
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to delete repository',
        variant: 'destructive',
      });
    }
  };

  const handleExpand = (repoId: number) => {
    if (expandedRepo === repoId) {
      setExpandedRepo(null);
    } else {
      setExpandedRepo(repoId);
      const repo = repositories.find(r => r.id === repoId);
      if (repo && !repo.branches) {
        fetchBranches(repoId);
      }
    }
  };

  const handleSync = async (repoId: number) => {
    setSyncingRepo(repoId);
    try {
      await syncRepository(repoId);
      toast({
        title: 'Success',
        description: 'Repository synchronized successfully',
      });
      // Refresh both repository list and branches
      await fetchRepositories();
      if (expandedRepo === repoId) {
        await fetchBranches(repoId);
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : String(error);
      toast({
        title: 'Error',
        description: errorMessage,
        variant: 'destructive',
      });
    } finally {
      setSyncingRepo(null);
    }
  };

  if (loading) {
    return <div>Loading...</div>;
  }

  return (
    <div className="container mx-auto py-10">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">Repositories</h1>
        <AddRepositoryDialog onRepositoryAdded={fetchRepositories} />
      </div>
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-[50px]"></TableHead>
              <TableHead>Name</TableHead>
              <TableHead>URL</TableHead>
              <TableHead>Description</TableHead>
              <TableHead>Last Updated</TableHead>
              <TableHead>Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {repositories.map((repo) => (
              <>
                <TableRow key={repo.id}>
                  <TableCell>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => handleExpand(repo.id)}
                      className="p-0 h-8 w-8"
                    >
                      <ChevronDown
                        className={cn(
                          "h-4 w-4 transition-transform",
                          expandedRepo === repo.id && "transform rotate-180"
                        )}
                      />
                    </Button>
                  </TableCell>
                  <TableCell>{repo.name}</TableCell>
                  <TableCell>{repo.url}</TableCell>
                  <TableCell>{repo.description}</TableCell>
                  <TableCell>
                    {formatDate(repo.lastUpdated).date}
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleSync(repo.id)}
                        disabled={syncingRepo === repo.id}
                        className="p-0 h-8 w-8"
                      >
                        <RotateCw className={cn(
                          "h-4 w-4",
                          syncingRepo === repo.id && "animate-spin"
                        )} />
                      </Button>
                      <Button
                        variant="destructive"
                        size="sm"
                        onClick={() => handleDelete(repo.id)}
                      >
                        Delete
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
                <TableRow>
                  <TableCell colSpan={6} className="p-0">
                    <Collapsible open={expandedRepo === repo.id}>
                      <CollapsibleContent>
                        <div className="p-4 bg-muted/50">
                          <div className="flex items-center gap-2 mb-6">
                            <GitBranch className="h-5 w-5" />
                            <h3 className="text-lg font-semibold">Branches</h3>
                          </div>
                          {repo.branches ? (
                            <div className="pl-6 space-y-6">
                              {(() => {
                                const localBranches = repo.branches.filter(b => !b.isRemote);
                                const remoteBranches = repo.branches.filter(b => b.isRemote);
                                return (
                                  <>
                                    {localBranches.length > 0 && (
                                      <BranchList
                                        branches={localBranches}
                                        title="Local Branches"
                                        icon={Laptop}
                                        repositoryId={repo.id}
                                        onBranchesChange={() => handleBranchesChange(repo.id)}
                                      />
                                    )}
                                    {remoteBranches.length > 0 && (
                                      <BranchList
                                        branches={remoteBranches}
                                        title="Remote Branches"
                                        icon={Globe}
                                        repositoryId={repo.id}
                                        onBranchesChange={() => handleBranchesChange(repo.id)}
                                      />
                                    )}
                                  </>
                                );
                              })()}
                            </div>
                          ) : (
                            <div className="text-sm text-muted-foreground">
                              Loading branches...
                            </div>
                          )}
                        </div>
                      </CollapsibleContent>
                    </Collapsible>
                  </TableCell>
                </TableRow>
              </>
            ))}
          </TableBody>
        </Table>
      </div>
    </div>
  );
} 