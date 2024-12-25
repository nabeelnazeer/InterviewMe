'use client';
import { useState } from 'react';
import { FiUpload, FiTrash2 } from 'react-icons/fi';

export default function CVScoring() {
  const [uploadedCV, setUploadedCV] = useState(null);
  const [jobDescription, setJobDescription] = useState('');
  const [scores, setScores] = useState(null);
  const [isUploading, setIsUploading] = useState(false);
  const [error, setError] = useState(null);

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

  return (
    <div className="min-h-screen bg-gray-900 p-8">
      <div className="max-w-6xl mx-auto">
        <h1 className="text-3xl font-bold text-white mb-8">CV Scoring Dashboard</h1>
        
        {/* Upload and Job Description Section */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-8">
          <div className="bg-gray-800 p-6 rounded-lg">
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
                onClick={() => setUploadedCV(null)}
                className="mt-4 flex items-center text-red-400 hover:text-red-300"
              >
                <FiTrash2 className="mr-2" /> Remove CV
              </button>
            )}
          </div>

          <div className="bg-gray-800 p-6 rounded-lg">
            <h2 className="text-xl text-white mb-4">Job Description</h2>
            <textarea
              className="w-full h-48 bg-gray-700 text-white rounded-lg p-4 focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="Enter job description here..."
              value={jobDescription}
              onChange={(e) => setJobDescription(e.target.value)}
            />
          </div>
        </div>

        {/* Scoring Section */}
        {scores && (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
            <ScoreCard title="Skills Match" score={scores.skillMatch} color="blue" />
            <ScoreCard title="Experience Match" score={scores.experienceMatch} color="green" />
            <ScoreCard title="Education Match" score={scores.educationMatch} color="purple" />
            <ScoreCard title="Overall Score" score={scores.overallScore} color="yellow" />
          </div>
        )}
      </div>
    </div>
  );
}

const ScoreCard = ({ title, score, color }) => {
  const colorClasses = {
    blue: 'from-blue-500 to-blue-600',
    green: 'from-green-500 to-green-600',
    purple: 'from-purple-500 to-purple-600',
    yellow: 'from-yellow-500 to-yellow-600'
  };

  return (
    <div className={`bg-gradient-to-br ${colorClasses[color]} p-6 rounded-lg`}>
      <h3 className="text-lg font-semibold text-white mb-2">{title}</h3>
      <div className="flex items-end">
        <span className="text-4xl font-bold text-white">{score}</span>
        <span className="text-white text-xl ml-1">%</span>
      </div>
    </div>
  );
};
