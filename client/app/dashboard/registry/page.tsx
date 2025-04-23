"use client"

import { useEffect, useState } from "react"
import { useRouter } from "next/navigation"
import { 
  Plus, 
  Pencil, 
  Trash2, 
  AlertCircle, 
  Check, 
  X,
  Box
} from "lucide-react"
import { toast } from "sonner"

import { Button } from "@/components/ui/button"
import { 
  Card, 
  CardContent, 
  CardDescription, 
  CardFooter, 
  CardHeader, 
  CardTitle 
} from "@/components/ui/card"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import { 
  createRegistry, 
  deleteRegistry, 
  listRegistries, 
  updateRegistry 
} from "@/lib/api"
import { Registry } from "@/lib/types"

export default function RegistryPage() {
  const router = useRouter()
  const [registries, setRegistries] = useState<Registry[]>([])
  const [loading, setLoading] = useState(true)
  const [openCreateDialog, setOpenCreateDialog] = useState(false)
  const [openEditDialog, setOpenEditDialog] = useState(false)
  const [openDeleteDialog, setOpenDeleteDialog] = useState(false)
  const [selectedRegistry, setSelectedRegistry] = useState<Registry | null>(null)
  const [formData, setFormData] = useState({
    name: "",
    url: "",
    username: "",
    password: "",
    description: "",
  })

  const fetchRegistries = async () => {
    try {
      setLoading(true)
      const data = await listRegistries()
      // Initialize isOnline as false for all registries
      const registriesWithStatus = data.map(registry => ({
        ...registry,
        isOnline: false
      }))
      setRegistries(registriesWithStatus)
    } catch (error) {
      toast.error("Failed to fetch registries")
      console.error(error)
    } finally {
      setLoading(false)
    }
  }

  const handleRegistryStatusChange = (registryId: number, isOnline: boolean) => {
    console.log(`Registry status changed - ID: ${registryId}, Online: ${isOnline}`);
    setRegistries(prevRegistries => {
      const updatedRegistries = prevRegistries.map(registry => 
        registry.id === registryId 
          ? { ...registry, isOnline } 
          : registry
      );
      console.log('Updated registries:', updatedRegistries);
      return updatedRegistries;
    });
  }

  useEffect(() => {
    fetchRegistries()
  }, [])

  const handleInputChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>
  ) => {
    const { name, value } = e.target
    setFormData((prev) => ({ ...prev, [name]: value }))
  }

  const resetForm = () => {
    setFormData({
      name: "",
      url: "",
      username: "",
      password: "",
      description: "",
    })
    setSelectedRegistry(null)
  }

  const handleCreateRegistry = async () => {
    try {
      await createRegistry({
        name: formData.name,
        url: formData.url,
        username: formData.username,
        password: formData.password,
        description: formData.description,
      })
      toast.success("Registry created successfully")
      setOpenCreateDialog(false)
      resetForm()
      fetchRegistries()
    } catch (error) {
      toast.error("Failed to create registry")
      console.error(error)
    }
  }

  const handleEditRegistry = (registry: Registry) => {
    setSelectedRegistry(registry)
    setFormData({
      name: registry.name,
      url: registry.url,
      username: registry.username || "",
      password: "", // Don't populate password for security reasons
      description: registry.description || "",
    })
    setOpenEditDialog(true)
  }

  const handleUpdateRegistry = async () => {
    if (!selectedRegistry) return

    try {
      await updateRegistry(selectedRegistry.id, {
        name: formData.name,
        url: formData.url,
        username: formData.username,
        password: formData.password, // Only sent if changed
        description: formData.description,
      })
      toast.success("Registry updated successfully")
      setOpenEditDialog(false)
      resetForm()
      fetchRegistries()
    } catch (error) {
      toast.error("Failed to update registry")
      console.error(error)
    }
  }

  const handleDeleteClick = (registry: Registry) => {
    setSelectedRegistry(registry)
    setOpenDeleteDialog(true)
  }

  const handleDeleteRegistry = async () => {
    if (!selectedRegistry) return

    try {
      await deleteRegistry(selectedRegistry.id)
      toast.success("Registry deleted successfully")
      setOpenDeleteDialog(false)
      setSelectedRegistry(null)
      fetchRegistries()
    } catch (error) {
      toast.error("Failed to delete registry")
      console.error(error)
    }
  }

  return (
    <div className="container mx-auto py-6">
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Registry Management</h1>
          <p className="text-muted-foreground">
            Manage your Docker image registries for builds
          </p>
        </div>
        <Dialog open={openCreateDialog} onOpenChange={setOpenCreateDialog}>
          <DialogTrigger asChild>
            <Button onClick={() => resetForm()}>
              <Plus className="mr-2 h-4 w-4" /> Add Registry
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Add New Registry</DialogTitle>
              <DialogDescription>
                Enter the details for the new Docker registry
              </DialogDescription>
            </DialogHeader>
            <div className="grid gap-4 py-4">
              <div className="grid gap-2">
                <Label htmlFor="name">Registry Name</Label>
                <Input
                  id="name"
                  name="name"
                  placeholder="My Registry"
                  value={formData.name}
                  onChange={handleInputChange}
                />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="url">Registry URL</Label>
                <Input
                  id="url"
                  name="url"
                  placeholder="https://registry.example.com"
                  value={formData.url}
                  onChange={handleInputChange}
                />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="username">Username (optional)</Label>
                <Input
                  id="username"
                  name="username"
                  placeholder="username"
                  value={formData.username}
                  onChange={handleInputChange}
                />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="password">Password (optional)</Label>
                <Input
                  id="password"
                  name="password"
                  type="password"
                  placeholder="password"
                  value={formData.password}
                  onChange={handleInputChange}
                />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="description">Description (optional)</Label>
                <textarea
                  id="description"
                  name="description"
                  placeholder="Description of this registry"
                  value={formData.description}
                  onChange={handleInputChange}
                  className="flex min-h-[80px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                />
              </div>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setOpenCreateDialog(false)}>
                Cancel
              </Button>
              <Button onClick={handleCreateRegistry}>Create Registry</Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Registries</CardTitle>
          <CardDescription>
            List of all configured Docker registries
          </CardDescription>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex justify-center py-8">
              <p>Loading registries...</p>
            </div>
          ) : registries.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-8 text-center">
              <AlertCircle className="h-10 w-10 text-muted-foreground mb-4" />
              <h3 className="text-lg font-semibold">No Registries Found</h3>
              <p className="text-muted-foreground">
                You haven't added any registries yet.
              </p>
            </div>
          ) : (
            <div className="overflow-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Name</TableHead>
                    <TableHead>URL</TableHead>
                    <TableHead>Description</TableHead>
                    <TableHead>Authentication</TableHead>
                    <TableHead className="text-right">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {registries.map((registry) => (
                    <TableRow 
                      key={registry.id}
                      className="cursor-pointer hover:bg-muted/50"
                      onClick={() => router.push(`/dashboard/registry/${registry.id}/images`)}
                    >
                      <TableCell className="font-medium">{registry.name}</TableCell>
                      <TableCell>{registry.url}</TableCell>
                      <TableCell>
                        {registry.description ? (
                          <span className="text-sm text-muted-foreground max-w-xs truncate">
                            {registry.description}
                          </span>
                        ) : (
                          <span className="text-sm text-muted-foreground">No description</span>
                        )}
                      </TableCell>
                      <TableCell>
                        {registry.username ? (
                          <span className="flex items-center">
                            <Check className="mr-2 h-4 w-4 text-green-500" />
                            Configured
                          </span>
                        ) : (
                          <span className="flex items-center">
                            <X className="mr-2 h-4 w-4 text-gray-500" />
                            None
                          </span>
                        )}
                      </TableCell>
                      <TableCell className="text-right">
                        <TooltipProvider>
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <Button
                                variant="ghost"
                                size="icon"
                                onClick={(e) => {
                                  e.stopPropagation();
                                  router.push(`/dashboard/registry/${registry.id}/images`);
                                }}
                              >
                                <Box className="h-4 w-4" />
                              </Button>
                            </TooltipTrigger>
                            <TooltipContent>
                              Manage Images
                            </TooltipContent>
                          </Tooltip>
                        </TooltipProvider>
                        <TooltipProvider>
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <Button
                                variant="ghost"
                                size="icon"
                                onClick={(e) => {
                                  e.stopPropagation();
                                  handleEditRegistry(registry);
                                }}
                              >
                                <Pencil className="h-4 w-4" />
                              </Button>
                            </TooltipTrigger>
                            <TooltipContent>Edit Registry</TooltipContent>
                          </Tooltip>
                        </TooltipProvider>
                        <TooltipProvider>
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <Button
                                variant="ghost"
                                size="icon"
                                onClick={(e) => {
                                  e.stopPropagation();
                                  handleDeleteClick(registry);
                                }}
                              >
                                <Trash2 className="h-4 w-4" />
                              </Button>
                            </TooltipTrigger>
                            <TooltipContent>Delete Registry</TooltipContent>
                          </Tooltip>
                        </TooltipProvider>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Edit Registry Dialog */}
      <Dialog open={openEditDialog} onOpenChange={setOpenEditDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit Registry</DialogTitle>
            <DialogDescription>
              Modify the details for this Docker registry
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="edit-name">Registry Name</Label>
              <Input
                id="edit-name"
                name="name"
                placeholder="My Registry"
                value={formData.name}
                onChange={handleInputChange}
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="edit-url">Registry URL</Label>
              <Input
                id="edit-url"
                name="url"
                placeholder="https://registry.example.com"
                value={formData.url}
                onChange={handleInputChange}
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="edit-username">Username (optional)</Label>
              <Input
                id="edit-username"
                name="username"
                placeholder="username"
                value={formData.username}
                onChange={handleInputChange}
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="edit-password">
                Password (leave empty to keep unchanged)
              </Label>
              <Input
                id="edit-password"
                name="password"
                type="password"
                placeholder="••••••••"
                value={formData.password}
                onChange={handleInputChange}
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="edit-description">Description (optional)</Label>
              <textarea
                id="edit-description"
                name="description"
                placeholder="Description of this registry"
                value={formData.description}
                onChange={handleInputChange}
                className="flex min-h-[80px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setOpenEditDialog(false)}>
              Cancel
            </Button>
            <Button onClick={handleUpdateRegistry}>Save Changes</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Registry Dialog */}
      <Dialog open={openDeleteDialog} onOpenChange={setOpenDeleteDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete Registry</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete this registry? This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setOpenDeleteDialog(false)}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={handleDeleteRegistry}>
              Delete
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
} 