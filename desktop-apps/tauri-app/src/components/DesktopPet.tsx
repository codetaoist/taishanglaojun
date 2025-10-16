import { useState, useEffect, useRef, useCallback, useMemo } from 'react';
import { 
  MessageCircle, 
  Volume2, 
  VolumeX, 
  Minimize2,
  X,
  Heart,
  Coffee,
  Zap,
  Moon,
  Sun
} from 'lucide-react';
import { cn } from '../utils/cn';

interface DesktopPetProps {
  visible: boolean;
  onClose: () => void;
  onMinimize: () => void;
}

interface PetState {
  mood: 'happy' | 'normal' | 'sleepy' | 'excited';
  energy: number;
  happiness: number;
  isAsleep: boolean;
}

interface ChatMessage {
  id: string;
  text: string;
  isUser: boolean;
  timestamp: Date;
}

export default function DesktopPet({ visible, onClose, onMinimize }: DesktopPetProps) {
  const [petState, setPetState] = useState<PetState>({
    mood: 'normal',
    energy: 80,
    happiness: 70,
    isAsleep: false
  });
  
  const [showChat, setShowChat] = useState(false);
  const [chatMessages, setChatMessages] = useState<ChatMessage[]>([
    {
      id: '1',
      text: '你好！我是你的桌面助手小君，有什么可以帮助你的吗？',
      isUser: false,
      timestamp: new Date()
    }
  ]);
  const [inputMessage, setInputMessage] = useState('');
  const [soundEnabled, setSoundEnabled] = useState(true);
  const [isDragging, setIsDragging] = useState(false);
  const [position, setPosition] = useState({ x: 100, y: 100 });
  const [dragOffset, setDragOffset] = useState({ x: 0, y: 0 });
  
  const petRef = useRef<HTMLDivElement>(null);
  const chatRef = useRef<HTMLDivElement>(null);

  // 桌宠动画状态
  const [animation, setAnimation] = useState('idle');
  const [currentExpression, setCurrentExpression] = useState('😊');

  // 表情映射 - 使用 useMemo 优化
  const expressions = useMemo(() => ({
    happy: ['😊', '😄', '🥰', '😍', '🌟', '💖'],
    normal: ['😊', '🙂', '😌', '🤔', '😇'],
    sleepy: ['😴', '🥱', '😪', '💤', '🌙'],
    excited: ['🤩', '😆', '🎉', '✨', '🚀', '⚡']
  }), []);

  // 随机消息回复 - 使用 useMemo 优化
  const responses = useMemo(() => [
    '我明白了！',
    '这很有趣呢！',
    '让我想想...',
    '好的，我会记住的！',
    '你说得对！',
    '还有什么想聊的吗？',
    '我很高兴能帮助你！',
    '这是个好问题！',
    '哇，真棒！',
    '我学到了新东西！',
    '继续聊吧～',
    '你真聪明！'
  ], []);

  // 自动更新桌宠状态
  useEffect(() => {
    const interval = setInterval(() => {
      setPetState(prev => {
        const newEnergy = Math.max(0, prev.energy - 1);
        const newHappiness = Math.max(0, prev.happiness - 0.5);
        
        let newMood = prev.mood;
        if (newEnergy < 20) {
          newMood = 'sleepy';
        } else if (newHappiness > 80) {
          newMood = 'happy';
        } else if (newHappiness < 30) {
          newMood = 'normal';
        }

        return {
          ...prev,
          energy: newEnergy,
          happiness: newHappiness,
          mood: newMood,
          isAsleep: newEnergy < 10
        };
      });
    }, 30000); // 每30秒更新一次

    return () => clearInterval(interval);
  }, []);

  // 更新表情
  useEffect(() => {
    const interval = setInterval(() => {
      const moodExpressions = expressions[petState.mood];
      const randomExpression = moodExpressions[Math.floor(Math.random() * moodExpressions.length)];
      setCurrentExpression(randomExpression);
    }, 3000); // 每3秒换一次表情

    return () => clearInterval(interval);
  }, [petState.mood]);

  // 拖拽功能
  const handleMouseDown = (e: React.MouseEvent) => {
    if (e.target === petRef.current || petRef.current?.contains(e.target as Node)) {
      setIsDragging(true);
      const rect = petRef.current!.getBoundingClientRect();
      setDragOffset({
        x: e.clientX - rect.left,
        y: e.clientY - rect.top
      });
    }
  };

  const handleMouseMove = (e: MouseEvent) => {
    if (isDragging) {
      setPosition({
        x: e.clientX - dragOffset.x,
        y: e.clientY - dragOffset.y
      });
    }
  };

  const handleMouseUp = () => {
    setIsDragging(false);
  };

  useEffect(() => {
    if (isDragging) {
      document.addEventListener('mousemove', handleMouseMove);
      document.addEventListener('mouseup', handleMouseUp);
    }

    return () => {
      document.removeEventListener('mousemove', handleMouseMove);
      document.removeEventListener('mouseup', handleMouseUp);
    };
  }, [isDragging, dragOffset]);

  // 发送消息 - 使用 useCallback 优化
  const handleSendMessage = useCallback(() => {
    if (!inputMessage.trim()) return;

    const userMessage: ChatMessage = {
      id: Date.now().toString(),
      text: inputMessage,
      isUser: true,
      timestamp: new Date()
    };

    setChatMessages(prev => [...prev, userMessage]);

    // 模拟AI回复
    setTimeout(() => {
      const response = responses[Math.floor(Math.random() * responses.length)];
      const aiMessage: ChatMessage = {
        id: (Date.now() + 1).toString(),
        text: response,
        isUser: false,
        timestamp: new Date()
      };
      setChatMessages(prev => [...prev, aiMessage]);
      
      // 增加快乐值
      setPetState(prev => ({
        ...prev,
        happiness: Math.min(100, prev.happiness + 5)
      }));
    }, 1000);

    setInputMessage('');
  }, [inputMessage, responses]);

  // 喂食功能 - 使用 useCallback 优化
  const handleFeed = useCallback(() => {
    setPetState(prev => ({
      ...prev,
      energy: Math.min(100, prev.energy + 20),
      happiness: Math.min(100, prev.happiness + 10),
      mood: 'happy'
    }));
    setAnimation('eating');
    setTimeout(() => setAnimation('idle'), 2000);
  }, []);

  // 玩耍功能 - 使用 useCallback 优化
  const handlePlay = useCallback(() => {
    setPetState(prev => ({
      ...prev,
      happiness: Math.min(100, prev.happiness + 15),
      energy: Math.max(0, prev.energy - 5),
      mood: 'excited'
    }));
    setAnimation('playing');
    setTimeout(() => setAnimation('idle'), 3000);
  }, []);

  // 睡觉功能 - 使用 useCallback 优化
  const handleSleep = useCallback(() => {
    setPetState(prev => {
      const newIsAsleep = !prev.isAsleep;
      setAnimation(newIsAsleep ? 'sleeping' : 'waking');
      setTimeout(() => setAnimation('idle'), 2000);
      
      return {
        ...prev,
        isAsleep: newIsAsleep,
        energy: newIsAsleep ? prev.energy : Math.min(100, prev.energy + 30),
        mood: newIsAsleep ? 'sleepy' : 'normal'
      };
    });
  }, []);

  if (!visible) return null;

  return (
    <div className="fixed inset-0 pointer-events-none z-50">
      {/* 桌宠主体 */}
      <div
        ref={petRef}
        className={cn(
          'absolute pointer-events-auto cursor-move select-none',
          'bg-white/90 backdrop-blur-sm rounded-2xl shadow-2xl border border-gray-200',
          'transition-all duration-300',
          isDragging && 'scale-105 shadow-3xl',
          petState.isAsleep && 'opacity-70'
        )}
        style={{
          left: position.x,
          top: position.y,
          width: '200px',
          height: '280px'
        }}
        onMouseDown={handleMouseDown}
      >
        {/* 控制按钮 */}
        <div className="absolute top-2 right-2 flex space-x-1">
          <button
            onClick={() => setSoundEnabled(!soundEnabled)}
            className="p-1 rounded-full bg-gray-100 hover:bg-gray-200 transition-colors"
          >
            {soundEnabled ? (
              <Volume2 className="h-3 w-3 text-gray-600" />
            ) : (
              <VolumeX className="h-3 w-3 text-gray-600" />
            )}
          </button>
          <button
            onClick={onMinimize}
            className="p-1 rounded-full bg-gray-100 hover:bg-gray-200 transition-colors"
          >
            <Minimize2 className="h-3 w-3 text-gray-600" />
          </button>
          <button
            onClick={onClose}
            className="p-1 rounded-full bg-red-100 hover:bg-red-200 transition-colors"
          >
            <X className="h-3 w-3 text-red-600" />
          </button>
        </div>

        {/* 桌宠形象 */}
        <div className="flex flex-col items-center justify-center h-full p-4">
          <div className={cn(
            'text-6xl mb-2 transition-all duration-500 ease-in-out transform',
            'hover:scale-110 cursor-pointer select-none',
            animation === 'bounce' && 'pet-bounce',
            animation === 'float' && 'pet-float',
            animation === 'wiggle' && 'pet-wiggle',
            animation === 'eating' && 'pet-wiggle',
            animation === 'playing' && 'pet-bounce',
            animation === 'sleeping' && 'pet-sleep',
            animation === 'waking' && 'pet-wiggle',
            petState.isAsleep && 'opacity-70 grayscale',
            petState.mood === 'happy' && 'drop-shadow-lg',
            petState.mood === 'excited' && 'animate-pulse'
          )}
          onClick={() => setAnimation('wiggle')}
          >
            {petState.isAsleep ? '😴' : currentExpression}
          </div>

          {/* 状态栏 */}
          <div className="flex space-x-3 mb-4">
            <div className="flex items-center space-x-2">
              <Heart className={cn(
                "h-4 w-4 transition-colors duration-300",
                petState.happiness > 70 ? "text-red-500" : 
                petState.happiness > 30 ? "text-orange-500" : "text-gray-400"
              )} />
              <div className="w-16 h-3 bg-gray-200 rounded-full overflow-hidden shadow-inner">
                <div 
                  className={cn(
                    "h-full transition-all duration-500 ease-out rounded-full",
                    petState.happiness > 70 ? "bg-gradient-to-r from-red-400 to-red-500" :
                    petState.happiness > 30 ? "bg-gradient-to-r from-orange-400 to-orange-500" :
                    "bg-gradient-to-r from-gray-400 to-gray-500"
                  )}
                  style={{ width: `${petState.happiness}%` }}
                />
              </div>
              <span className="text-xs text-gray-600 font-medium">{Math.round(petState.happiness)}</span>
            </div>
            <div className="flex items-center space-x-2">
              <Zap className={cn(
                "h-4 w-4 transition-colors duration-300",
                petState.energy > 70 ? "text-yellow-500" : 
                petState.energy > 30 ? "text-orange-500" : "text-red-500"
              )} />
              <div className="w-16 h-3 bg-gray-200 rounded-full overflow-hidden shadow-inner">
                <div 
                  className={cn(
                    "h-full transition-all duration-500 ease-out rounded-full",
                    petState.energy > 70 ? "bg-gradient-to-r from-yellow-400 to-yellow-500" :
                    petState.energy > 30 ? "bg-gradient-to-r from-orange-400 to-orange-500" :
                    "bg-gradient-to-r from-red-400 to-red-500"
                  )}
                  style={{ width: `${petState.energy}%` }}
                />
              </div>
              <span className="text-xs text-gray-600 font-medium">{Math.round(petState.energy)}</span>
            </div>
          </div>

          {/* 交互按钮 */}
          <div className="flex space-x-2 justify-center">
            <button
              onClick={handleFeed}
              className={cn(
                "p-3 rounded-xl transition-all duration-300 transform hover:scale-110 active:scale-95",
                "bg-gradient-to-br from-green-100 to-green-200 hover:from-green-200 hover:to-green-300",
                "shadow-md hover:shadow-lg border border-green-200",
                "disabled:opacity-50 disabled:cursor-not-allowed"
              )}
              title="喂食 (+20 能量, +10 快乐)"
              disabled={petState.energy >= 100}
            >
              <Coffee className="h-4 w-4 text-green-700" />
            </button>
            <button
              onClick={handlePlay}
              className={cn(
                "p-3 rounded-xl transition-all duration-300 transform hover:scale-110 active:scale-95",
                "bg-gradient-to-br from-yellow-100 to-yellow-200 hover:from-yellow-200 hover:to-yellow-300",
                "shadow-md hover:shadow-lg border border-yellow-200",
                "disabled:opacity-50 disabled:cursor-not-allowed"
              )}
              title="玩耍 (+15 快乐, -5 能量)"
              disabled={petState.energy < 10}
            >
              <Zap className="h-4 w-4 text-yellow-700" />
            </button>
            <button
              onClick={handleSleep}
              className={cn(
                "p-3 rounded-xl transition-all duration-300 transform hover:scale-110 active:scale-95",
                "bg-gradient-to-br from-purple-100 to-purple-200 hover:from-purple-200 hover:to-purple-300",
                "shadow-md hover:shadow-lg border border-purple-200"
              )}
              title={petState.isAsleep ? "唤醒" : "睡觉 (+30 能量)"}
            >
              {petState.isAsleep ? (
                <Sun className="h-4 w-4 text-purple-700" />
              ) : (
                <Moon className="h-4 w-4 text-purple-700" />
              )}
            </button>
            <button
              onClick={() => setShowChat(!showChat)}
              className={cn(
                "p-3 rounded-xl transition-all duration-300 transform hover:scale-110 active:scale-95",
                "bg-gradient-to-br from-blue-100 to-blue-200 hover:from-blue-200 hover:to-blue-300",
                "shadow-md hover:shadow-lg border border-blue-200",
                showChat && "ring-2 ring-blue-400 ring-opacity-50"
              )}
              title="聊天"
            >
              <MessageCircle className="h-4 w-4 text-blue-700" />
            </button>
          </div>
        </div>
      </div>

      {/* 聊天窗口 */}
      {showChat && (
        <div
          ref={chatRef}
          className="absolute pointer-events-auto bg-gradient-to-br from-white to-gray-50 rounded-xl shadow-2xl border border-gray-200 animate-in slide-in-from-right-2 duration-300"
          style={{
            left: position.x + 220,
            top: position.y,
            width: '300px',
            height: '400px'
          }}
        >
          <div className="flex items-center justify-between p-4 border-b border-gray-200 bg-gradient-to-r from-blue-50 to-purple-50">
            <h3 className="font-semibold text-gray-800 flex items-center space-x-2">
              <MessageCircle className="h-4 w-4 text-blue-600" />
              <span>与小君聊天</span>
            </h3>
            <button
              onClick={() => setShowChat(false)}
              className="p-1.5 rounded-full hover:bg-white/80 transition-all duration-200 hover:scale-110"
            >
              <X className="h-4 w-4 text-gray-600" />
            </button>
          </div>

          <div className="flex-1 p-4 overflow-y-auto scrollbar-thin scrollbar-thumb-gray-300" style={{ height: '300px' }}>
            <div className="space-y-4">
              {chatMessages.map((message) => (
                <div
                  key={message.id}
                  className={cn(
                    'flex animate-in slide-in-from-bottom-2 duration-300',
                    message.isUser ? 'justify-end' : 'justify-start'
                  )}
                >
                  <div
                    className={cn(
                      'max-w-[80%] p-3 rounded-2xl text-sm shadow-sm transition-all duration-200 hover:shadow-md',
                      message.isUser
                        ? 'bg-gradient-to-r from-blue-500 to-blue-600 text-white'
                        : 'bg-gradient-to-r from-gray-100 to-gray-200 text-gray-800 border border-gray-200'
                    )}
                  >
                    {message.text}
                  </div>
                </div>
              ))}
              {chatMessages.length <= 1 && (
                <div className="text-center text-gray-400 text-sm py-8">
                  开始和你的宠物聊天吧！ 🐾
                </div>
              )}
            </div>
          </div>

          <div className="p-4 border-t border-gray-200 bg-gradient-to-r from-gray-50 to-white">
            <div className="flex space-x-3">
              <input
                type="text"
                value={inputMessage}
                onChange={(e) => setInputMessage(e.target.value)}
                onKeyPress={(e) => e.key === 'Enter' && handleSendMessage()}
                placeholder="输入消息..."
                className="flex-1 px-4 py-3 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-400 focus:border-transparent transition-all duration-200 bg-white shadow-sm text-sm"
              />
              <button
                onClick={handleSendMessage}
                disabled={!inputMessage.trim()}
                className={cn(
                  "px-4 py-3 rounded-xl transition-all duration-200 transform hover:scale-105 active:scale-95 shadow-md text-sm font-medium",
                  "bg-gradient-to-r from-blue-500 to-blue-600 hover:from-blue-600 hover:to-blue-700",
                  "text-white disabled:opacity-50 disabled:cursor-not-allowed disabled:transform-none"
                )}
              >
                发送
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}