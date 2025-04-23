import { NextRequest, NextResponse } from 'next/server';

// Environment variables
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

// Create a registry
export async function POST(request: NextRequest) {
  try {
    const body = await request.json();
    
    const response = await fetch(`${API_BASE_URL}/registries`, {
      method: 'POST',
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
      return NextResponse.json({ error: 'Failed to create registry' }, { 
        status: response.status,
        headers: {
          'Content-Type': 'application/json'
        }
      });
    }
    
    const data = await response.json();
    
    return NextResponse.json(data, {
      status: 201,
      headers: {
        'Content-Type': 'application/json'
      }
    });
  } catch (error) {
    console.error('Error creating registry:', error);
    return NextResponse.json({ error: 'Failed to create registry' }, { status: 500 });
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