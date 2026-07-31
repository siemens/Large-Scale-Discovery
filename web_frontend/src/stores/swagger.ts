/*
 * Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
 *
 * Copyright (c) Siemens AG, 2025-2026.
 *
 * This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
 * directory or visit <https://opensource.org/licenses/MIT>.
 *
 * SPDX-License-Identifier: MIT
 */

import { defineStore } from 'pinia'
import { ref } from 'vue'

import http from '@/api/http'
import { useAsyncState } from '@/composables/useAsyncState'
import { type EndpointGroup, parseSwagger, type SwaggerDoc } from '@/lib/swagger'
import { SWAGGER_CONFIG } from '@/lib/swaggerConfig'

/** Owns the parsed Swagger/OpenAPI documentation. */
export const useSwaggerStore = defineStore('swagger', () => {
  // Store the API title shown above the documentation
  const title = ref('')

  // Store the advertised backend version
  const version = ref('')

  // Store the introductory API description
  const description = ref('')

  // Store the common path prepended to documented endpoints
  const basePath = ref('')

  // Store parsed endpoints in the groups rendered by the API page
  const groups = ref<EndpointGroup[]>([])

  // Remember whether the Swagger document has been loaded
  const isLoaded = ref(false)

  // Track the document request for loading and error feedback
  const fetchState = useAsyncState()

  // Load and parse the Swagger document
  async function fetchDocs(): Promise<void> {
    await fetchState.run(async () => {
      // Fetch the document from backends with either wrapped or direct responses
      const response = await http.get<SwaggerDoc & { body?: SwaggerDoc }>('/swagger/doc.json')

      // Unwrap backends that place the Swagger document inside a body property
      const document = response.data?.body ?? response.data

      // Convert the raw document into the small model used by the UI
      const parsedDocument = parseSwagger(document, SWAGGER_CONFIG)

      // Copy document metadata into focused reactive fields
      title.value = parsedDocument.title
      version.value = parsedDocument.version
      description.value = parsedDocument.description
      basePath.value = parsedDocument.basePath

      // Publish endpoints only after parsing has completed
      groups.value = parsedDocument.groups
      isLoaded.value = true
    })
  }

  // Clear the parsed document when application state must be discarded
  function reset(): void {
    // Reset document metadata to its initial values
    title.value = ''
    version.value = ''
    description.value = ''
    basePath.value = ''

    // Reset endpoint data and its request lifecycle
    groups.value = []
    isLoaded.value = false
    fetchState.reset()
  }

  // Expose parsed document data together with its request lifecycle
  return {
    title,
    version,
    description,
    basePath,
    groups,
    isLoaded,
    isFetching: fetchState.isLoading,
    fetchError: fetchState.error,
    fetchDocs,
    reset,
  }
})
