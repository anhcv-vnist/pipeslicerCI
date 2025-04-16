'use client';

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { ArrowRight, GitBranch, Box, Settings, Shield } from "lucide-react";
import Link from "next/link";

export default function Home() {
  return (
    <div className="min-h-screen bg-background">
      {/* Hero Section */}
      <section className="py-20 px-4 sm:px-6 lg:px-8">
        <div className="max-w-7xl mx-auto text-center">
          <h1 className="text-4xl sm:text-6xl font-bold tracking-tight">
            PipeSlicer CI
          </h1>
          <p className="mt-6 text-xl text-muted-foreground max-w-3xl mx-auto">
            Streamline your CI/CD pipeline with our powerful and intuitive continuous integration platform.
          </p>
          <div className="mt-10 flex gap-4 justify-center">
            <Link href="/repositories">
              <Button size="lg" className="gap-2">
                Get Started <ArrowRight className="h-4 w-4" />
              </Button>
            </Link>
            <Button variant="outline" size="lg">
              Learn More
            </Button>
          </div>
        </div>
      </section>

      {/* Features Section */}
      <section className="py-16 px-4 sm:px-6 lg:px-8 bg-muted/50">
        <div className="max-w-7xl mx-auto">
          <h2 className="text-3xl font-bold text-center mb-12">Key Features</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            <Card>
              <CardHeader>
                <GitBranch className="h-8 w-8 mb-2 text-primary" />
                <CardTitle>Repository Management</CardTitle>
              </CardHeader>
              <CardContent>
                <CardDescription>
                  Easily connect and manage your Git repositories. Track changes and automate your workflow.
                </CardDescription>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <Box className="h-8 w-8 mb-2 text-primary" />
                <CardTitle>Image Building</CardTitle>
              </CardHeader>
              <CardContent>
                <CardDescription>
                  Build and manage container images with automated pipelines and version control.
                </CardDescription>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <Settings className="h-8 w-8 mb-2 text-primary" />
                <CardTitle>Pipeline Configuration</CardTitle>
              </CardHeader>
              <CardContent>
                <CardDescription>
                  Configure and customize your CI/CD pipelines with an intuitive interface.
                </CardDescription>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <Shield className="h-8 w-8 mb-2 text-primary" />
                <CardTitle>Registry Integration</CardTitle>
              </CardHeader>
              <CardContent>
                <CardDescription>
                  Seamlessly integrate with container registries for secure image storage and distribution.
                </CardDescription>
              </CardContent>
            </Card>
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="py-20 px-4 sm:px-6 lg:px-8">
        <div className="max-w-4xl mx-auto text-center">
          <h2 className="text-3xl font-bold mb-6">Ready to Get Started?</h2>
          <p className="text-xl text-muted-foreground mb-8">
            Join the growing number of teams using PipeSlicer CI to streamline their development workflow.
          </p>
          <Link href="/repositories">
            <Button size="lg" className="gap-2">
              Start Managing Repositories <ArrowRight className="h-4 w-4" />
            </Button>
          </Link>
        </div>
      </section>
    </div>
  );
} 