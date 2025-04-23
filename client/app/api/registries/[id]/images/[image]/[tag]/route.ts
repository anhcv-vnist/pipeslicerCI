import { NextRequest, NextResponse } from 'next/server';

// Environment variables
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

// Get Docker image details
export async function GET(
  request: NextRequest, 
  { params }: { params: { id: string; image: string; tag: string } }
) {
  try {
    const { id, image, tag } = params;
    const timestamp = new Date().getTime();
    const response = await fetch(
      `${API_BASE_URL}/registries/${id}/images/${encodeURIComponent(image)}/${encodeURIComponent(tag)}?_=${timestamp}`, 
      {
        cache: 'no-store',
        headers: {
          'Cache-Control': 'no-cache, no-store, must-revalidate',
          'Pragma': 'no-cache',
          'Expires': '0'
        }
      }
    );
    
    if (!response.ok) {
      console.error('Error response from API:', response.status, response.statusText);
      return NextResponse.json({ error: 'Failed to fetch image details' }, { 
        status: response.status,
        headers: {
          'Content-Type': 'application/json',
          'Cache-Control': 'no-cache, no-store, must-revalidate',
          'Pragma': 'no-cache',
          'Expires': '0'
        }
      });
    }
    
    const data = await response.json();
    
    return new NextResponse(JSON.stringify(data), {
      headers: {
        'Content-Type': 'application/json',
        'Cache-Control': 'no-cache, no-store, must-revalidate',
        'Pragma': 'no-cache',
        'Expires': '0'
      }
    });
  } catch (error) {
    console.error('Error fetching image details:', error);
    return NextResponse.json({ error: 'Failed to fetch image details' }, { 
      status: 500,
      headers: {
        'Content-Type': 'application/json',
        'Cache-Control': 'no-cache, no-store, must-revalidate',
        'Pragma': 'no-cache',
        'Expires': '0'
      } 
    });
  }
}

// Delete Docker image
export async function DELETE(
  request: NextRequest, 
  { params }: { params: { id: string; image: string; tag: string } }
) {
  try {
    const { id, image, tag } = params;
    const response = await fetch(
      `${API_BASE_URL}/registries/${id}/images/${encodeURIComponent(image)}/${encodeURIComponent(tag)}`, 
      {
        method: 'DELETE',
        headers: {
          'Cache-Control': 'no-cache, no-store, must-revalidate',
          'Pragma': 'no-cache',
          'Expires': '0'
        }
      }
    );
    
    if (!response.ok) {
      console.error('Error response from API:', response.status, response.statusText);
      return NextResponse.json({ error: 'Failed to delete image' }, { 
        status: response.status,
        headers: {
          'Content-Type': 'application/json'
        }
      });
    }
    
    return new NextResponse(null, { status: 204 });
  } catch (error) {
    console.error('Error deleting image:', error);
    return NextResponse.json({ error: 'Failed to delete image' }, { status: 500 });
  }
}

// Handle preflight requests
export async function OPTIONS() {
  return new NextResponse(null, {
    status: 204,
    headers: {
      'Access-Control-Allow-Origin': '*',
      'Access-Control-Allow-Methods': 'GET, POST, PUT, DELETE, OPTIONS',
      'Access-Control-Allow-Headers': 'Content-Type, Authorization',
      'Access-Control-Max-Age': '86400'
    }
  });
} 