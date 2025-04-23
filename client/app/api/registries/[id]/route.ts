import { NextRequest, NextResponse } from 'next/server';

// Environment variables
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

// Get a single registry
export async function GET(request: NextRequest, { params }: { params: { id: string } }) {
  try {
    const id = params.id;
    const timestamp = new Date().getTime();
    const response = await fetch(`${API_BASE_URL}/registries/${id}?_=${timestamp}`, {
      cache: 'no-store',
      headers: {
        'Cache-Control': 'no-cache, no-store, must-revalidate',
        'Pragma': 'no-cache',
        'Expires': '0'
      }
    });
    
    if (!response.ok) {
      console.error('Error response from API:', response.status, response.statusText);
      return NextResponse.json({ error: 'Failed to fetch registry' }, { 
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
    console.error('Error fetching registry:', error);
    return NextResponse.json({ error: 'Failed to fetch registry' }, { 
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

// Update a registry
export async function PUT(request: NextRequest, { params }: { params: { id: string } }) {
  try {
    const id = params.id;
    const body = await request.json();
    
    const response = await fetch(`${API_BASE_URL}/registries/${id}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'Cache-Control': 'no-cache, no-store, must-revalidate',
        'Pragma': 'no-cache',
        'Expires': '0'
      },
      body: JSON.stringify(body)
    });
    
    if (!response.ok) {
      console.error('Error response from API:', response.status, response.statusText);
      return NextResponse.json({ error: 'Failed to update registry' }, { 
        status: response.status,
        headers: {
          'Content-Type': 'application/json'
        }
      });
    }
    
    const data = await response.json();
    
    return NextResponse.json(data);
  } catch (error) {
    console.error('Error updating registry:', error);
    return NextResponse.json({ error: 'Failed to update registry' }, { status: 500 });
  }
}

// Delete a registry
export async function DELETE(request: NextRequest, { params }: { params: { id: string } }) {
  try {
    const id = params.id;
    
    const response = await fetch(`${API_BASE_URL}/registries/${id}`, {
      method: 'DELETE',
      headers: {
        'Cache-Control': 'no-cache, no-store, must-revalidate',
        'Pragma': 'no-cache',
        'Expires': '0'
      }
    });
    
    if (!response.ok) {
      console.error('Error response from API:', response.status, response.statusText);
      return NextResponse.json({ error: 'Failed to delete registry' }, { 
        status: response.status,
        headers: {
          'Content-Type': 'application/json'
        }
      });
    }
    
    return new NextResponse(null, { status: 204 });
  } catch (error) {
    console.error('Error deleting registry:', error);
    return NextResponse.json({ error: 'Failed to delete registry' }, { status: 500 });
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