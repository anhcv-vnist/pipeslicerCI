"use client"

import { useEffect, useState } from "react"
import { useRouter } from "next/navigation"
import { 
  ArrowLeft, 
  RefreshCw, 
  Trash2, 
  Tag, 
  Copy, 
  AlertCircle, 
  Check, 
  X,
  Info
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
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { 
  getRegistry, 
  listRegistryImages, 
  getImageDetail, 
  deleteImage, 
  retagImage, 
  copyImage,
  testRegistryConnection
} from "@/lib/api"
import { Registry, DockerImage, DockerImageDetail, RetagImageRequest, CopyImageRequest } from "@/lib/types"
import { formatBytes, formatDate } from "@/lib/utils"

export default function RegistryImagesPage({ params }: { params: { id: string } }) {
  const router = useRouter()
  const registryId = parseInt(params.id)
  const [registry, setRegistry] = useState<Registry | null>(null)
  const [images, setImages] = useState<DockerImage[]>([])
  const [loading, setLoading] = useState(true)
  const [selectedImage, setSelectedImage] = useState<DockerImage | null>(null)
  const [imageDetail, setImageDetail] = useState<DockerImageDetail | null>(null)
  const [openDeleteDialog, setOpenDeleteDialog] = useState(false)
  const [openRetagDialog, setOpenRetagDialog] = useState(false)
  const [openCopyDialog, setOpenCopyDialog] = useState(false)
  const [openDetailDialog, setOpenDetailDialog] = useState(false)
  const [retagForm, setRetagForm] = useState<RetagImageRequest>({
    source_image: "",
    source_tag: "",
    destination_image: "",
    destination_tag: ""
  })
  const [copyForm, setCopyForm] = useState<CopyImageRequest>({
    sourceRegistryId: 0,
    sourceImage: "",
    sourceTag: "",
    destinationRegistryId: 0,
    destinationImage: "",
    destinationTag: ""
  })
  const [registries, setRegistries] = useState<Registry[]>([])

  useEffect(() => {
    fetchRegistry()
  }, [registryId])

  const fetchRegistry = async () => {
    try {
      setLoading(true)
      const data = await getRegistry(registryId)
      setRegistry(data)
      
      // Test registry connection
      try {
        const connectionStatus = await testRegistryConnection(registryId)
        if (connectionStatus.status === 'success') {
          // Registry is online, fetch images
          await fetchImages()
        } else {
          toast.warning("Registry is offline. Images may not be available.")
          // Still try to fetch images, but show a warning
          await fetchImages()
        }
      } catch (error) {
        console.error("Failed to test registry connection:", error)
        toast.warning("Could not verify registry connection. Images may not be available.")
        // Still try to fetch images
        await fetchImages()
      }
    } catch (error) {
      toast.error("Failed to fetch registry")
      console.error(error)
      router.push('/dashboard/registry')
    } finally {
      setLoading(false)
    }
  }

  const fetchImages = async () => {
    try {
      const data = await listRegistryImages(registryId)
      console.log("Fetched images:", data);
      setImages(data)
    } catch (error) {
      console.error("Failed to fetch registry images:", error)
      toast.error("Failed to fetch registry images. The registry may be offline or unreachable.")
      // Set empty images array to show the "No images found" message
      setImages([])
    }
  }

  const handleImageClick = async (image: DockerImage) => {
    setSelectedImage(image)
    try {
      const detail = await getImageDetail(registryId, image.name, image.tags?.[0] || '')
      setImageDetail(detail)
      setOpenDetailDialog(true)
    } catch (error) {
      toast.error("Failed to fetch image details")
      console.error(error)
    }
  }

  const handleDeleteClick = (image: DockerImage) => {
    setSelectedImage(image)
    setOpenDeleteDialog(true)
  }

  const handleDeleteImage = async () => {
    if (!selectedImage) return

    try {
      const tag = selectedImage.tags?.[0] || '';
      if (!tag) {
        toast.info("Deleting image with default 'latest' tag");
      }
      
      await deleteImage(registryId, selectedImage.name, tag);
      toast.success("Image deleted successfully");
      setOpenDeleteDialog(false);
      setSelectedImage(null);
      fetchImages();
    } catch (error) {
      toast.error("Failed to delete image");
      console.error(error);
    }
  }

  const handleRetagClick = (image: DockerImage) => {
    setSelectedImage(image)
    setRetagForm({
      source_image: image.name,
      source_tag: image.tags?.[0] || '',
      destination_image: image.name,
      destination_tag: ""
    })
    setOpenRetagDialog(true)
  }

  const handleRetagImage = async () => {
    if (!selectedImage) return

    try {
      await retagImage(registryId, retagForm)
      toast.success("Image retagged successfully")
      setOpenRetagDialog(false)
      setSelectedImage(null)
      fetchImages()
    } catch (error) {
      toast.error("Failed to retag image")
      console.error(error)
    }
  }

  const handleCopyClick = (image: DockerImage) => {
    setSelectedImage(image)
    setCopyForm({
      sourceRegistryId: registryId,
      sourceImage: image.name,
      sourceTag: image.tags?.[0] || '',
      destinationRegistryId: 0,
      destinationImage: image.name,
      destinationTag: image.tags?.[0] || ''
    })
    setOpenCopyDialog(true)
  }

  const handleCopyImage = async () => {
    if (!selectedImage) return

    try {
      await copyImage(copyForm)
      toast.success("Image copied successfully")
      setOpenCopyDialog(false)
      setSelectedImage(null)
    } catch (error) {
      toast.error("Failed to copy image")
      console.error(error)
    }
  }

  const handleInputChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>
  ) => {
    const { name, value } = e.target
    if (name.startsWith('retag_')) {
      const field = name.replace('retag_', '')
      setRetagForm(prev => ({ ...prev, [field]: value }))
    } else if (name.startsWith('copy_')) {
      const field = name.replace('copy_', '')
      setCopyForm(prev => ({ ...prev, [field]: value }))
    }
  }

  const handleSelectChange = (name: string, value: string) => {
    if (name === 'copy_destinationRegistryId') {
      setCopyForm(prev => ({ ...prev, destinationRegistryId: parseInt(value) }))
    }
  }

  if (loading) {
    return (
      <div className="container mx-auto py-6">
        <div className="flex items-center space-x-4 mb-6">
          <Button variant="ghost" onClick={() => router.push('/dashboard/registry')}>
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back to Registries
          </Button>
        </div>
        <Card>
          <CardHeader>
            <CardTitle>Loading...</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-center p-8">
              <RefreshCw className="h-6 w-6 animate-spin" />
            </div>
          </CardContent>
        </Card>
      </div>
    )
  }

  if (!registry) {
    return (
      <div className="container mx-auto py-6">
        <div className="flex items-center space-x-4 mb-6">
          <Button variant="ghost" onClick={() => router.push('/dashboard/registry')}>
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back to Registries
          </Button>
        </div>
        <Card>
          <CardHeader>
            <CardTitle>Registry Not Found</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-center p-8">
              <AlertCircle className="h-6 w-6 text-red-500 mr-2" />
              <span>The requested registry could not be found.</span>
            </div>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="container mx-auto py-6">
      <div className="flex items-center mb-6">
        <Button 
          variant="outline" 
          size="icon" 
          className="mr-2"
          onClick={() => router.push('/dashboard/registry')}
        >
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <div>
          <h1 className="text-3xl font-bold tracking-tight">
            {registry?.name || 'Registry'} Images
          </h1>
          <p className="text-muted-foreground">
            Manage Docker images in this registry
          </p>
        </div>
      </div>

      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <div>
            <CardTitle>Docker Images</CardTitle>
            <CardDescription>
              List of Docker images in {registry?.name || 'this registry'}
            </CardDescription>
          </div>
          <Button onClick={fetchImages} variant="outline" size="sm">
            <RefreshCw className="h-4 w-4 mr-2" />
            Refresh
          </Button>
        </CardHeader>
        <CardContent>
          {images.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-8 text-center">
              <AlertCircle className="h-10 w-10 text-muted-foreground mb-4" />
              <h3 className="text-lg font-semibold">No Images Found</h3>
              <p className="text-muted-foreground">
                There are no Docker images in this registry.
              </p>
            </div>
          ) : (
            <div className="overflow-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Name</TableHead>
                    <TableHead>Tags</TableHead>
                    <TableHead>Size</TableHead>
                    <TableHead>Created</TableHead>
                    <TableHead>Last Updated</TableHead>
                    <TableHead className="text-right">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {images.map((image) => (
                    <TableRow key={`${image.name}-${image.tags?.[0] || 'untagged'}`}>
                      <TableCell className="font-medium">{image.name}</TableCell>
                      <TableCell>
                        <div className="flex flex-wrap gap-1">
                          {image.tags && image.tags.length > 0 ? (
                            image.tags.map((tag) => (
                              <span 
                                key={tag} 
                                className="bg-muted px-2 py-1 rounded-md text-xs"
                              >
                                {tag}
                              </span>
                            ))
                          ) : (
                            <span className="text-muted-foreground text-xs">untagged</span>
                          )}
                        </div>
                      </TableCell>
                      <TableCell>{formatBytes(image.size)}</TableCell>
                      <TableCell>{formatDate(image.createdAt)}</TableCell>
                      <TableCell>{formatDate(image.lastUpdated)}</TableCell>
                      <TableCell className="text-right">
                        <TooltipProvider>
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <Button
                                variant="ghost"
                                size="icon"
                                onClick={() => handleImageClick(image)}
                              >
                                <Info className="h-4 w-4" />
                              </Button>
                            </TooltipTrigger>
                            <TooltipContent>View Details</TooltipContent>
                          </Tooltip>
                        </TooltipProvider>
                        <TooltipProvider>
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <Button
                                variant="ghost"
                                size="icon"
                                onClick={() => handleRetagClick(image)}
                              >
                                <Tag className="h-4 w-4" />
                              </Button>
                            </TooltipTrigger>
                            <TooltipContent>Retag Image</TooltipContent>
                          </Tooltip>
                        </TooltipProvider>
                        <TooltipProvider>
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <Button
                                variant="ghost"
                                size="icon"
                                onClick={() => handleCopyClick(image)}
                              >
                                <Copy className="h-4 w-4" />
                              </Button>
                            </TooltipTrigger>
                            <TooltipContent>Copy Image</TooltipContent>
                          </Tooltip>
                        </TooltipProvider>
                        <TooltipProvider>
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <Button
                                variant="ghost"
                                size="icon"
                                onClick={() => handleDeleteClick(image)}
                              >
                                <Trash2 className="h-4 w-4" />
                              </Button>
                            </TooltipTrigger>
                            <TooltipContent>Delete Image</TooltipContent>
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

      {/* Image Detail Dialog */}
      <Dialog open={openDetailDialog} onOpenChange={setOpenDetailDialog}>
        <DialogContent className="max-w-4xl max-h-[80vh] overflow-hidden flex flex-col">
          <DialogHeader>
            <DialogTitle>Image Details</DialogTitle>
            <DialogDescription>
              Detailed information about {selectedImage?.name}:{selectedImage?.tags?.[0] || 'untagged'}
            </DialogDescription>
          </DialogHeader>
          {imageDetail && (
            <div className="overflow-y-auto pr-2 flex-1">
              <div className="grid gap-4 py-4">
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <h3 className="font-semibold mb-2">Basic Information</h3>
                    <div className="space-y-2">
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">Name:</span>
                        <span>{imageDetail.name}</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">Tags:</span>
                        <div className="flex flex-wrap gap-1">
                          {imageDetail.tags.map((tag) => (
                            <span 
                              key={tag} 
                              className="bg-muted px-2 py-1 rounded-md text-xs"
                            >
                              {tag}
                            </span>
                          ))}
                        </div>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">Size:</span>
                        <span>{formatBytes(imageDetail.size)}</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">Created:</span>
                        <span>{formatDate(imageDetail.createdAt)}</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">Last Updated:</span>
                        <span>{formatDate(imageDetail.lastUpdated)}</span>
                      </div>
                    </div>
                  </div>
                  <div>
                    <h3 className="font-semibold mb-2">Configuration</h3>
                    <div className="space-y-2">
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">Architecture:</span>
                        <span>{imageDetail.config.architecture}</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">OS:</span>
                        <span>{imageDetail.config.os}</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">Environment:</span>
                        <div className="flex flex-col">
                          {imageDetail.config.env.map((env, index) => (
                            <span key={index} className="text-xs">{env}</span>
                          ))}
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
                
                <div>
                  <h3 className="font-semibold mb-2">Layers</h3>
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>Digest</TableHead>
                        <TableHead>Size</TableHead>
                        <TableHead>Created</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {imageDetail.layers.map((layer, index) => (
                        <TableRow key={index}>
                          <TableCell className="font-mono text-xs">
                            {layer.digest.substring(0, 20)}...
                          </TableCell>
                          <TableCell>{formatBytes(layer.size)}</TableCell>
                          <TableCell>{formatDate(layer.createdAt)}</TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>
                
                <div>
                  <h3 className="font-semibold mb-2">History</h3>
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>Created</TableHead>
                        <TableHead>Created By</TableHead>
                        <TableHead>Comment</TableHead>
                        <TableHead>Empty Layer</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {imageDetail.history.map((history, index) => (
                        <TableRow key={index}>
                          <TableCell>{formatDate(history.created)}</TableCell>
                          <TableCell className="font-mono text-xs">
                            {history.createdBy}
                          </TableCell>
                          <TableCell>{history.comment}</TableCell>
                          <TableCell>
                            {history.emptyLayer ? (
                              <Check className="h-4 w-4 text-green-500" />
                            ) : (
                              <X className="h-4 w-4 text-red-500" />
                            )}
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>
              </div>
            </div>
          )}
          <DialogFooter className="mt-4">
            <Button onClick={() => setOpenDetailDialog(false)}>Close</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Image Dialog */}
      <Dialog open={openDeleteDialog} onOpenChange={setOpenDeleteDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete Image</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete {selectedImage?.name}:{selectedImage?.tags?.[0] || 'untagged'}? This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setOpenDeleteDialog(false)}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={handleDeleteImage}>
              Delete
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Retag Image Dialog */}
      <Dialog open={openRetagDialog} onOpenChange={setOpenRetagDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Retag Image</DialogTitle>
            <DialogDescription>
              Create a new tag for {selectedImage?.name}:{selectedImage?.tags?.[0] || 'untagged'}
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="retag_source_image">Source Image</Label>
              <Input
                id="retag_source_image"
                name="retag_source_image"
                value={retagForm.source_image}
                onChange={handleInputChange}
                disabled
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="retag_source_tag">Source Tag</Label>
              <Input
                id="retag_source_tag"
                name="retag_source_tag"
                value={retagForm.source_tag}
                onChange={handleInputChange}
                disabled
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="retag_destination_image">Destination Image</Label>
              <Input
                id="retag_destination_image"
                name="retag_destination_image"
                placeholder="e.g. myapp"
                value={retagForm.destination_image}
                onChange={handleInputChange}
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="retag_destination_tag">Destination Tag</Label>
              <Input
                id="retag_destination_tag"
                name="retag_destination_tag"
                placeholder="e.g. v1.0.0"
                value={retagForm.destination_tag}
                onChange={handleInputChange}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setOpenRetagDialog(false)}>
              Cancel
            </Button>
            <Button onClick={handleRetagImage}>Retag</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Copy Image Dialog */}
      <Dialog open={openCopyDialog} onOpenChange={setOpenCopyDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Copy Image</DialogTitle>
            <DialogDescription>
              Copy {selectedImage?.name}:{selectedImage?.tags?.[0] || 'untagged'} to another registry
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="copy_sourceRegistryId">Source Registry</Label>
              <Input
                id="copy_sourceRegistryId"
                name="copy_sourceRegistryId"
                value={copyForm.sourceRegistryId}
                disabled
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="copy_sourceImage">Source Image</Label>
              <Input
                id="copy_sourceImage"
                name="copy_sourceImage"
                value={copyForm.sourceImage}
                disabled
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="copy_sourceTag">Source Tag</Label>
              <Input
                id="copy_sourceTag"
                name="copy_sourceTag"
                value={copyForm.sourceTag}
                disabled
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="copy_destinationRegistryId">Destination Registry</Label>
              <Select
                value={copyForm.destinationRegistryId.toString()}
                onValueChange={(value) => handleSelectChange('copy_destinationRegistryId', value)}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Select a registry" />
                </SelectTrigger>
                <SelectContent>
                  {registries.map((registry) => (
                    <SelectItem key={registry.id} value={registry.id.toString()}>
                      {registry.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="grid gap-2">
              <Label htmlFor="copy_destinationImage">Destination Image</Label>
              <Input
                id="copy_destinationImage"
                name="copy_destinationImage"
                placeholder="e.g. myapp"
                value={copyForm.destinationImage}
                onChange={handleInputChange}
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="copy_destinationTag">Destination Tag</Label>
              <Input
                id="copy_destinationTag"
                name="copy_destinationTag"
                placeholder="e.g. v1.0.0"
                value={copyForm.destinationTag}
                onChange={handleInputChange}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setOpenCopyDialog(false)}>
              Cancel
            </Button>
            <Button onClick={handleCopyImage}>Copy</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
} 