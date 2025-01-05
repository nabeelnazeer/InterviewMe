'use client';
import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import dynamic from 'next/dynamic';
import axios from 'axios';
import { FaCheckCircle, FaBriefcase, FaGraduationCap, FaTools, FaChartLine, 
         FaLanguage, FaCertificate, FaProjectDiagram, FaUserTie } from 'react-icons/fa';

const AnalysisPage = () => {
  const [isMounted, setIsMounted] = useState(false);
  const router = useRouter();
  const [error, setError] = useState(null);
  const [showContent, setShowContent] = useState(false);
  const [analysisData, setAnalysisData] = useState(null);
  const [pdfData, setPdfData] = useState(null);

  useEffect(() => {
    setIsMounted(true);
  }, []);

  useEffect(() => {
    if (!isMounted) return;
    
    try {
      const storedData = localStorage.getItem('scoringResults');
      if (!storedData) {
        throw new Error('No scoring data found');
      }

      const parsedData = JSON.parse(storedData);
      const { scores, resumeData } = parsedData;

      if (!scores || !resumeData) {
        throw new Error('Invalid scoring data');
      }

      // Save the filename in the analysis data
      setAnalysisData({
        overall_score: Math.round(scores.overall_score || 0),
        technical_score: Math.round(scores.detailed_scores?.technical_skills || 0),
        soft_skills_score: Math.round(scores.detailed_scores?.soft_skills || 0),
        experience_score: Math.round(scores.experience_match || 0),
        education_score: Math.round(scores.education_match || 0),
        education: resumeData.entities.education || [],
        experience: resumeData.entities.experience || [],
        projects: resumeData.entities.projects || [],
        skills: resumeData.entities.skills || [],
        recommendations: scores.feedback || ['No recommendations available'],
        matchedSkills: scores.matched_skills || { exact_matches: [], partial_matches: [] },
        soft_skills_analysis: scores.soft_skills_analysis || {
          score: 0,
          extracted_skills: [],
          experience_based_skills: []
        },
        filename: resumeData.filename, // Add this line
      });

      setShowContent(true);

    } catch (error) {
      console.error('Error:', error);
      setError('Unable to load analysis data');
      setTimeout(() => router.push('/cvScoring'), 3000);
    }
  }, [isMounted, router]);

  const handlePdfUpload = async (file) => {
    try {
      const formData = new FormData();
      formData.append('file', file);

      const response = await axios.post('http://localhost:8080/api/upload', formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      });

      if (response.data.filename) {
        setPdfData(`http://localhost:8080/api/pdf/display?filename=${response.data.filename}`);
      }
    } catch (error) {
      console.error('Error uploading PDF:', error);
    }
  };

  const handleViewPDF = () => {
    try {
      // Get filename from stored data
      const storedData = localStorage.getItem('scoringResults');
      if (!storedData) {
        console.error('No resume data found');
        return;
      }

      const parsedData = JSON.parse(storedData);
      const filename = parsedData.resumeData?.filename;
      
      if (filename) {
        // Just use the filename as stored, which should already include the "upload-" prefix
        setPdfData(`http://localhost:8080/api/pdf/display?filename=${filename}`);
      } else {
        console.error('No filename found in stored data');
      }
    } catch (error) {
      console.error('Error viewing PDF:', error);
    }
  };

  if (!isMounted) return null;

  if (error) {
    return (
      <div className="min-h-screen bg-gray-900 text-white p-8 flex items-center justify-center">
        <div className="bg-red-500/10 border border-red-500 rounded-lg p-4">
          {error}
        </div>
      </div>
    );
  }

  const EducationSection = ({ education }) => (
    <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
      <h3 className="text-xl font-bold mb-4 flex items-center">
        <FaGraduationCap className="mr-2 text-green-400" />
        Education
      </h3>
      <div className="space-y-4">
        {education.map((edu, index) => (
          <div key={index} className="p-4 bg-gray-700 rounded">
            <h4 className="font-semibold text-green-400">{edu.degree}</h4>
            <p className="text-gray-300">{edu.institution}</p>
            <p className="text-gray-400">{edu.year}</p>
            {edu.specialization && (
              <p className="text-gray-300 mt-1">Specialization: {edu.specialization}</p>
            )}
          </div>
        ))}
      </div>
    </div>
  );

  const ExperienceSection = ({ experience }) => (
    <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
      <h3 className="text-xl font-bold mb-4 flex items-center">
        <FaBriefcase className="mr-2 text-green-400" />
        Experience
      </h3>
      <div className="space-y-4">
        {experience.map((exp, index) => (
          <div key={index} className="p-4 bg-gray-700 rounded">
            <h4 className="font-semibold text-green-400">{exp.title || exp.position}</h4>
            <p className="text-gray-300">{exp.company}</p>
            <p className="text-gray-400">{exp.duration}</p>
            {exp.description && (
              <p className="text-gray-300 mt-2 text-sm">{exp.description}</p>
            )}
          </div>
        ))}
      </div>
    </div>
  );

  const ProjectsSection = ({ projects }) => (
    <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
      <h3 className="text-xl font-bold mb-4 flex items-center">
        <FaProjectDiagram className="mr-2 text-green-400" />
        Projects
      </h3>
      <div className="space-y-4">
        {projects.map((project, index) => (
          <div key={index} className="p-4 bg-gray-700 rounded">
            <h4 className="font-semibold text-green-400">{project.name}</h4>
            <p className="text-gray-300 text-sm mt-1">{project.description}</p>
            {project.technologies && project.technologies.length > 0 && (
              <div className="mt-2 flex flex-wrap gap-2">
                {project.technologies.map((tech, i) => (
                  <span key={i} className="px-2 py-1 bg-gray-600 rounded-full text-xs text-gray-300">
                    {tech}
                  </span>
                ))}
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  );

  const TechnicalSkillsSection = ({ score, matchedSkills }) => (
    <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
      <h3 className="text-xl font-bold mb-4 flex items-center">
        <FaTools className="mr-2 text-green-400" />
        Technical Skills
      </h3>
      <div className="text-4xl font-bold text-green-400 mb-6">
        {score}%
      </div>
      
      {/* Exact Matches */}
      {matchedSkills?.exact_matches?.length > 0 && (
        <div className="mb-6">
          <h4 className="text-lg font-semibold text-green-400 mb-3">Matched Skills With Model</h4>
          <div className="flex flex-wrap gap-2">
            {matchedSkills.exact_matches.map((skill, index) => (
              <span key={index} className="px-3 py-1 bg-green-500/20 border border-green-500 rounded-full text-sm">
                {skill}
              </span>
            ))}
          </div>
        </div>
      )}
    </div>
  );

  const SoftSkillsSection = ({ score, softSkillsAnalysis }) => (
    <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
      <h3 className="text-xl font-bold mb-4 flex items-center">
        <FaChartLine className="mr-2 text-green-400" />
        Soft Skills
      </h3>
      <div className="text-4xl font-bold text-green-400 mb-6">
        {score}%
      </div>
      
      {/* Extracted Soft Skills */}
      {softSkillsAnalysis?.extracted_skills?.length > 0 && (
        <div className="mb-6">
          <h4 className="text-lg font-semibold text-blue-400 mb-3">Identified Skills With Model</h4>
          <div className="flex flex-wrap gap-2">
            {softSkillsAnalysis.extracted_skills.map((skill, index) => (
              <span key={index} className="px-3 py-1 bg-blue-500/20 border border-blue-500 rounded-full text-sm">
                {skill}
              </span>
            ))}
          </div>
        </div>
      )}

      {/* Experience-based Skills */}
      {softSkillsAnalysis?.experience_based_skills?.length > 0 && (
        <div>
          <h4 className="text-lg font-semibold text-purple-400 mb-3">Experience-Based Skills</h4>
          <div className="flex flex-wrap gap-2">
            {softSkillsAnalysis.experience_based_skills.map((skill, index) => (
              <span key={index} className="px-3 py-1 bg-purple-500/20 border border-purple-500 rounded-full text-sm">
                {skill}
              </span>
            ))}
          </div>
        </div>
      )}
    </div>
  );

  return (
    <div className="min-h-screen bg-gray-900 text-white p-8">
      <div className="fixed top-4 right-4 z-50 flex space-x-4">
        <button
          onClick={handleViewPDF}
          className="px-6 py-3 bg-gray-800/90 hover:bg-blue-600 text-white rounded-lg 
                    font-semibold transition-all duration-300 flex items-center space-x-2
                    border border-blue-500 backdrop-blur-sm cursor-pointer"
        >
          <span>View Current Resume</span>
        </button>
        <button
          onClick={() => router.push('/CVScoring')}
          className="px-6 py-3 bg-gray-800/90 hover:bg-green-600 text-white rounded-lg 
                    font-semibold transition-all duration-300 flex items-center space-x-2
                    border border-green-500 backdrop-blur-sm"
        >
          <span>↩ Analyze Another Resume</span>
        </button>
      </div>

      {/* Add PDF viewer modal */}
      {pdfData && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-75 p-4">
          <div className="relative w-full max-w-4xl h-[80vh] bg-white rounded-lg">
            <button
              onClick={() => setPdfData(null)}
              className="absolute -top-2 -right-2 w-8 h-8 flex items-center justify-center 
                       bg-red-500 hover:bg-red-600 text-white rounded-full 
                       shadow-lg transform transition-all duration-200 
                       hover:scale-110 focus:outline-none"
              aria-label="Close PDF viewer"
            >
              ✕
            </button>
            <iframe
              src={pdfData}
              className="w-full h-full rounded-lg"
              title="PDF Viewer"
            />
          </div>
        </div>
      )}

      <div className="max-w-6xl mx-auto">
        {showContent && analysisData && (
          <div className="space-y-8">
            {/* Overall Score */}
            <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
              <h2 className="text-2xl font-bold mb-4 text-green-400">Overall Match Score</h2>
              <div className="text-6xl font-bold text-center text-green-400">
                {analysisData.overall_score}%
              </div>
            </div>

            {/* Detailed Scores */}
            <div className="grid md:grid-cols-2 gap-6">
              <TechnicalSkillsSection 
                score={analysisData.technical_score} 
                matchedSkills={analysisData.matchedSkills}
              />
              <SoftSkillsSection 
                score={analysisData.soft_skills_score}
                softSkillsAnalysis={analysisData.soft_skills_analysis}
              />
              <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
                <h3 className="text-xl font-bold mb-4 flex items-center">
                  <FaBriefcase className="mr-2 text-green-400" />
                  Experience
                </h3>
                <div className="text-4xl font-bold text-green-400">
                  {analysisData.experience_score}%
                </div>
              </div>
              <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
                <h3 className="text-xl font-bold mb-4 flex items-center">
                  <FaGraduationCap className="mr-2 text-green-400" />
                  Education
                </h3>
                <div className="text-4xl font-bold text-green-400">
                  {analysisData.education_score}%
                </div>
              </div>
            </div>

            <div className="grid md:grid-cols-2 gap-6">
              <EducationSection education={analysisData.education} />
              <ExperienceSection experience={analysisData.experience} />
            </div>

            <ProjectsSection projects={analysisData.projects} />

            {/* Recommendations */}
            <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
              <h3 className="text-xl font-bold mb-4 text-green-400">Recommendations</h3>
              <div className="space-y-3">
                {analysisData.recommendations.map((rec, index) => (
                  <div key={index} className="p-3 bg-gray-700 rounded flex items-center">
                    <FaCheckCircle className="text-green-400 mr-3" />
                    <span>{rec}</span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default dynamic(() => Promise.resolve(AnalysisPage), {
  ssr: false
});
