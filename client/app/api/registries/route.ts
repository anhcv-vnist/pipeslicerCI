import { NextResponse } from 'next/server';

// Environment variables
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export async function GET() {
  try {
    // Fetch from the backend API with cache-busting param
    const timestamp = new Date().getTime();
    const response = await fetch(`${API_BASE_URL}/registries?_=${timestamp}`, {
      cache: 'no-store',
      headers: {
        'Cache-Control': 'no-cache, no-store, must-revalidate',
        'Pragma': 'no-cache',
        'Expires': '0'
      }
    });
    
    if (!response.ok) {
      console.error('Error response from API:', response.status, response.statusText);
      return NextResponse.json({ error: 'Failed to fetch from API' }, { 
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
    
    console.log('API Response:', data);
    
    // Check what type of data we're getting
    let formattedData = data;
    
    if (!Array.isArray(data)) {
      if (data && typeof data === 'object') {
        // Check if the data is nested in a property
        if (data.registries && Array.isArray(data.registries)) {
          formattedData = data.registries;
        } else {
          // It's a single object, put it in an array
          formattedData = [data];
        }
      } else {
        // Not an object, return empty array
        formattedData = [];
      }
    }
    
    // Make sure all objects have properly cased keys
    // This is just for debugging
    if (Array.isArray(formattedData)) {
      formattedData = formattedData.map(item => {
        console.log('Registry item:', item);
        return item;
      });
    }
    
    // Return the processed data with no-cache headers
    return new NextResponse(JSON.stringify(formattedData), {
      headers: {
        'Content-Type': 'application/json',
        'Cache-Control': 'no-cache, no-store, must-revalidate',
        'Pragma': 'no-cache',
        'Expires': '0',
        'Access-Control-Allow-Origin': '*',
        'Access-Control-Allow-Methods': 'GET, POST, PUT, DELETE, OPTIONS',
        'Access-Control-Allow-Headers': 'Content-Type, Authorization'
      }
    });
  } catch (error) {
    console.error('Error fetching registries:', error);
    return NextResponse.json({ error: 'Failed to fetch registries' }, { 
      status: 500,
      headers: {
        'Content-Type': 'application/json',
        'Cache-Control': 'no-cache, no-store, must-revalidate',
        'Pragma': 'no-cache',
        'Expires': '0',
        'Access-Control-Allow-Origin': '*',
        'Access-Control-Allow-Methods': 'GET, POST, PUT, DELETE, OPTIONS',
        'Access-Control-Allow-Headers': 'Content-Type, Authorization'
      } 
    });
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