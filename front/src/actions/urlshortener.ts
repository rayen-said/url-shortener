// src/app/actions.ts
"use server"; // This directive marks all exports in this file as Server Actions

interface ShortenApiResponse {
  short_url?: string;
  error?: string; // Go service might return errors differently, adjust as needed
}

interface ActionResult {
  shortUrl?: string;
  error?: string;
}

// Ensure your Go backend is running and accessible, usually on http://localhost:8080


export async function shortenUrlAction(longUrl: string): Promise<ActionResult> {
  if (!longUrl || !longUrl.trim()) {
    return { error: "URL cannot be empty." };
  }

  // Basic client-side validation (can be more robust)
  if (!longUrl.startsWith("http://") && !longUrl.startsWith("https://")) {
    return { error: "Invalid URL. Must start with http:// or https://." };
  }

  try {
    if (!process.env.RAIL_PUBLIC_DOMAIN) {
      throw new Error("RAIL_PUBLIC_DOMAIN environment variable is not defined");
    }
    const response = await fetch(process.env.RAIL_PUBLIC_DOMAIN, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ url: longUrl }),
      cache: 'no-store', // Important for dynamic requests to API routes
    });

    if (!response.ok) {
      // Try to parse error from Go service if possible
      let errorMessage = `Error: ${response.status} ${response.statusText}`;
      try {
        const errorBody = await response.json();
        if (errorBody && typeof errorBody === 'string') { // Go error is plain text
            errorMessage = errorBody;
        } else if (errorBody && errorBody.error) { // Or if it's JSON with an error field
            errorMessage = errorBody.error;
        }
      } catch (e) {
        // Failed to parse error body, stick with status text
        console.error("Failed to parse error response from Go service:", e);
      }
      console.error("Backend error:", errorMessage);
      return { error: errorMessage };
    }

    const data: ShortenApiResponse = await response.json();

    if (data.error) {
      return { error: data.error };
    }

    if (!data.short_url) {
        return { error: "Failed to shorten URL. Backend response did not include a short URL." };
    }

    return { shortUrl: data.short_url };

  } catch (error: unknown) {
    console.error("Network or unexpected error in server action:", error);
    
      
    return { error: "An unexpected error occurred on the server. Please try again." };
  }
}