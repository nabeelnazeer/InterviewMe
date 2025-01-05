'use client';
import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import dynamic from 'next/dynamic';
import { FaCheckCircle, FaBriefcase, FaGraduationCap, FaTools, FaChartLine, 
         FaLanguage, FaCertificate, FaProjectDiagram, FaUserTie } from 'react-icons/fa';

const AnalysisPage = () => {
  const [isMounted, setIsMounted] = useState(false);
  const router = useRouter();
  const [error, setError] = useState(null);
  const [showContent, setShowContent] = useState(false);
  const [analysisData, setAnalysisData] = useState(null);

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
        recommendations: scores.feedback || ['No recommendations available']
      });

      setShowContent(true);

    } catch (error) {
      console.error('Error:', error);
      setError('Unable to load analysis data');
      setTimeout(() => router.push('/cvScoring'), 3000);
    }
  }, [isMounted, router]);

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

  return (
    <div className="min-h-screen bg-gray-900 text-white p-8">
      <div className="fixed top-4 right-4 z-50">
        <button
          onClick={() => router.push('/cvScoring')}
          className="px-6 py-3 bg-gray-800/90 hover:bg-green-600 text-white rounded-lg 
                    font-semibold transition-all duration-300 flex items-center space-x-2
                    border border-green-500 backdrop-blur-sm"
        >
          <span>↩ Analyze Another Resume</span>
        </button>
      </div>

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
              {[
                { icon: FaTools, title: "Technical Skills", score: analysisData.technical_score },
                { icon: FaChartLine, title: "Soft Skills", score: analysisData.soft_skills_score },
                { icon: FaBriefcase, title: "Experience", score: analysisData.experience_score },
                { icon: FaGraduationCap, title: "Education", score: analysisData.education_score }
              ].map((item, index) => (
                <div key={index} className="bg-gray-800 rounded-xl p-6 border border-gray-700">
                  <h3 className="text-xl font-bold mb-4 flex items-center">
                    <item.icon className="mr-2 text-green-400" />
                    {item.title}
                  </h3>
                  <div className="text-4xl font-bold text-green-400">
                    {item.score}%
                  </div>
                </div>
              ))}
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
