import { NextRequest, NextResponse } from 'next/server';

// Environment variables
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

// Copy Docker image between registries
export async function POST(request: NextRequest) {
  try {
    const body = await request.json();
    
    const response = await fetch(`${API_BASE_URL}/registries/images/copy`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Cache-Control': 'no-cache, no-store, must-revalidate',
        'Pragma': 'no-cache',
        'Expires': '0'
      },
      body: JSON.stringify({
        source_registry_id: body.sourceRegistryId,
        source_image: body.sourceImage,
        source_tag: body.sourceTag,
        destination_registry_id: body.destinationRegistryId,
        destination_image: body.destinationImage,
        destination_tag: body.destinationTag
      })
    });
    
    if (!response.ok) {
      console.error('Error response from API:', response.status, response.statusText);
      const errorText = await response.text();
      console.error('Error response body:', errorText);
      return NextResponse.json({ error: 'Failed to copy image' }, { 
        status: response.status,
        headers: {
          'Content-Type': 'application/json'
        }
      });
    }
    
    return new NextResponse(null, { status: 204 });
  } catch (error) {
    console.error('Error copying image:', error);
    return NextResponse.json({ error: 'Failed to copy image' }, { status: 500 });
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