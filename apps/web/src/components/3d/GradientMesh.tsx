'use client';

import React from 'react';

/**
 * GradientMesh — Smooth ambient gradient orbs background
 * Replaces Three.js particles with performant CSS-only gradients
 */
export default function GradientMesh() {
  return (
    <>
      <div className="gradient-mesh">
        <div className="gradient-orb gradient-orb-1" />
        <div className="gradient-orb gradient-orb-2" />
        <div className="gradient-orb gradient-orb-3" />
      </div>
      <div className="grain-overlay" />
    </>
  );
}
