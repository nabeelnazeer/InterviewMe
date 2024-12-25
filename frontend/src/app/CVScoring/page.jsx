'use client';
import { useState } from 'react';
import { FiUpload, FiTrash2, FiCpu } from 'react-icons/fi';

export default function CVScoring() {
  const [uploadedCV, setUploadedCV] = useState(null);
  const [jobDescription, setJobDescription] = useState('');
  const [modelScores, setModelScores] = useState({});
  const [isUploading, setIsUploading] = useState(false);
  const [error, setError] = useState(null);
  const [processingModels, setProcessingModels] = useState({});

  const aiModels = [
    { id: 'gpt4', name: 'GPT-4 Analysis', color: '#10B981' },
    { id: 'bert', name: 'BERT Matcher', color: '#6366F1' },
    { id: 'llama', name: 'LLaMA Evaluator', color: '#EC4899' },
    { id: 'custom', name: 'Custom AI Model', color: '#F59E0B' }
  ];

  const handleFileUpload = async (event) => {
    const file = event.target.files[0];
    if (file) {
      setIsUploading(true);
      setError(null);

      try {
        const formData = new FormData();
        formData.append('file', file);

        const response = await fetch('http://localhost:8080/api/upload', {
          method: 'POST',
          body: formData,
        });

        if (!response.ok) {
          throw new Error('Upload failed');
        }

        setUploadedCV(file);
        // TODO: Update scores based on API response
        setScores({
          skillMatch: 85,
          experienceMatch: 78,
          educationMatch: 92,
          overallScore: 85
        });
      } catch (err) {
        setError('Failed to upload CV. Please try again.');
        console.error('Upload error:', err);
      } finally {
        setIsUploading(false);
      }
    }
  };

  const handleDelete = async () => {
    try {
      const response = await fetch(`http://localhost:8080/api/delete?filename=${uploadedCV.name}`, {
        method: 'DELETE',
      });

      if (!response.ok) {
        throw new Error('Delete failed');
      }

      setUploadedCV(null);
      setScores(null);
    } catch (err) {
      setError('Failed to delete CV. Please try again.');
      console.error('Delete error:', err);
    }
  };

  const processWithModel = async (modelId) => {
    setProcessingModels(prev => ({ ...prev, [modelId]: true }));
    try {
      // Simulate API call to specific model endpoint
      const response = await fetch(`http://localhost:8080/api/analyze/${modelId}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ cv: uploadedCV?.name, jobDescription }),
      });
      
      if (!response.ok) throw new Error(`${modelId} analysis failed`);
      
      const result = await response.json();
      setModelScores(prev => ({ ...prev, [modelId]: result }));
    } catch (err) {
      setError(`Failed to process with ${modelId}. Please try again.`);
    } finally {
      setProcessingModels(prev => ({ ...prev, [modelId]: false }));
    }
  };

  return (
    <div className="min-h-screen bg-gray-900 p-8">
      <div className="max-w-7xl mx-auto">
        <h1 className="text-4xl font-bold text-white mb-8 background-blur">InterviewMe- CV-Score Module</h1>
        
        {/* Upload and Job Description Section */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-8">
          <div className="bg-gray-800 p-6 rounded-lg border-2 border-opacity-50" style={{ borderColor: '#10B981' }}>
            <h2 className="text-xl text-white mb-4">Upload CV</h2>
            <div className="flex items-center justify-center w-full">
              <label className={`flex flex-col items-center justify-center w-full h-48 border-2 border-gray-600 border-dashed rounded-lg cursor-pointer hover:bg-gray-700 ${isUploading ? 'opacity-50' : ''}`}>
                <div className="flex flex-col items-center justify-center pt-5 pb-6">
                  <FiUpload className="w-10 h-10 text-gray-400 mb-3" />
                  <p className="text-sm text-gray-400">
                    {isUploading ? "Uploading..." : uploadedCV ? uploadedCV.name : "Click to upload CV"}
                  </p>
                </div>
                <input 
                  type="file" 
                  className="hidden" 
                  onChange={handleFileUpload} 
                  accept=".pdf,.doc,.docx"
                  disabled={isUploading} 
                />
              </label>
            </div>
            {error && (
              <p className="mt-2 text-red-400 text-sm">{error}</p>
            )}
            {uploadedCV && (
              <button
                onClick={handleDelete}
                className="mt-4 flex items-center text-red-400 hover:text-red-300"
              >
                <FiTrash2 className="mr-2" /> Remove CV
              </button>
            )}
          </div>

          <div className="bg-gray-800 p-6 rounded-lg border-2 border-opacity-50" style={{ borderColor: '#6366F1' }}>
            <h2 className="text-xl text-white mb-4">Job Description</h2>
            <textarea
              className="w-full h-48 bg-gray-700 text-white rounded-lg p-4 focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="Enter job description here..."
              value={jobDescription}
              onChange={(e) => setJobDescription(e.target.value)}
            />
          </div>
        </div>

        {/* AI Models Section */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          {aiModels.map((model) => (
            <div
              key={model.id}
              className="bg-gray-800 p-6 rounded-lg border-2 border-opacity-50 transition-all hover:border-opacity-100"
              style={{ borderColor: model.color }}
            >
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-xl font-semibold text-white flex items-center">
                  <FiCpu className="mr-2" style={{ color: model.color }} />
                  {model.name}
                </h3>
              </div>

              {uploadedCV && (
                <button
                  onClick={() => processWithModel(model.id)}
                  disabled={processingModels[model.id]}
                  className="w-full py-2 px-4 rounded-lg text-white font-medium transition-all"
                  style={{
                    backgroundColor: processingModels[model.id] ? '#374151' : model.color,
                    opacity: processingModels[model.id] ? 0.7 : 1
                  }}
                >
                  {processingModels[model.id] ? 'Processing...' : 'Analyze CV'}
                </button>
              )}

              {modelScores[model.id] && (
                <div className="mt-4 space-y-2">
                  {Object.entries(modelScores[model.id]).map(([key, value]) => (
                    <div key={key} className="flex justify-between items-center">
                      <span className="text-gray-300 capitalize">{key.replace('_', ' ')}</span>
                      <span className="text-white font-semibold">{value}%</span>
                    </div>
                  ))}
                </div>
              )}
            </div>
          ))}
        </div>

        {error && (
          <div className="mt-4 p-4 bg-red-900/50 border border-red-500 rounded-lg text-red-200">
            {error}
          </div>
        )}
      </div>
    </div>
  );
}
