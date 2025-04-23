'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Registry } from '@/lib/types';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Pencil, Trash2, Package, Box } from 'lucide-react';
import {
  Dialog,
  DialogTrigger,
} from '@/components/ui/dialog';
import {
  AlertDialog,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog';
import { RegistryStatus } from './registry-status';
import { cn } from '@/lib/utils';
import { toast } from "sonner";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip"

interface RegistryListProps {
  registries: Registry[];
  isEditDialogOpen: boolean;
  isDeleteDialogOpen: boolean;
  selectedRegistry: Registry | null;
  onOpenEditDialog: (registry: Registry) => void;
  onOpenDeleteDialog: (registry: Registry) => void;
  setIsEditDialogOpen: (open: boolean) => void;
  setIsDeleteDialogOpen: (open: boolean) => void;
  onRegistryStatusChange: (registryId: number, isOnline: boolean) => void;
}

export function RegistryList({
  registries,
  isEditDialogOpen,
  isDeleteDialogOpen,
  selectedRegistry,
  onOpenEditDialog,
  onOpenDeleteDialog,
  setIsEditDialogOpen,
  setIsDeleteDialogOpen,
  onRegistryStatusChange,
}: RegistryListProps) {
  const router = useRouter();

  const handleRegistryClick = (registry: Registry) => {
    console.log('Registry clicked:', registry);
    console.log('Registry online status:', registry.isOnline);
    
    if (!registry.isOnline) {
      toast.error("Cannot access images: Registry is offline");
      return;
    }
    
    router.push(`/dashboard/registry/${registry.id}/images`);
  };

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      {registries.map((registry) => (
        <Card 
          key={registry.id || Math.random()}
          className={cn(
            'transition-opacity',
            !registry.isOnline && 'opacity-50',
            'cursor-pointer hover:shadow-md' // Always show cursor pointer
          )}
          onClick={() => handleRegistryClick(registry)}
        >
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle className="flex items-center">
                <Package className="h-5 w-5 mr-2" />
                {registry.name || 'Unnamed Registry'}
              </CardTitle>
              <RegistryStatus 
                registryId={registry.id} 
                onStatusChange={(isOnline) => onRegistryStatusChange(registry.id, isOnline)}
              />
            </div>
            <CardDescription>{registry.url || 'No URL'}</CardDescription>
          </CardHeader>
          <CardContent>
            {registry.description && (
              <p className="text-sm text-muted-foreground">
                {registry.description}
              </p>
            )}
          </CardContent>
          <CardFooter className="flex justify-end gap-2">
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button 
                    variant="outline" 
                    size="sm" 
                    onClick={(e) => {
                      e.stopPropagation();
                      router.push(`/dashboard/registry/${registry.id}/images`);
                    }}
                    disabled={!registry.isOnline}
                  >
                    <Box className="h-4 w-4 mr-1" />
                    Manage
                  </Button>
                </TooltipTrigger>
                <TooltipContent>
                  {registry.isOnline ? 'Manage Images' : 'Registry is offline'}
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>

            <Dialog 
              open={isEditDialogOpen && selectedRegistry?.id === registry.id} 
              onOpenChange={setIsEditDialogOpen}
            >
              <DialogTrigger asChild>
                <Button 
                  variant="outline" 
                  size="sm" 
                  onClick={(e) => {
                    e.stopPropagation();
                    onOpenEditDialog(registry);
                  }}
                  disabled={!registry.isOnline}
                >
                  <Pencil className="h-4 w-4 mr-1" />
                  Edit
                </Button>
              </DialogTrigger>
            </Dialog>

            <AlertDialog 
              open={isDeleteDialogOpen && selectedRegistry?.id === registry.id} 
              onOpenChange={setIsDeleteDialogOpen}
            >
              <AlertDialogTrigger asChild>
                <Button 
                  variant="destructive" 
                  size="sm" 
                  onClick={(e) => {
                    e.stopPropagation();
                    onOpenDeleteDialog(registry);
                  }}
                  disabled={!registry.isOnline}
                >
                  <Trash2 className="h-4 w-4 mr-1" />
                  Delete
                </Button>
              </AlertDialogTrigger>
            </AlertDialog>
          </CardFooter>
        </Card>
      ))}
    </div>
  );
} 