package usecase

import (
	"context"
	"log"
	"time"

	"mindfs/server/internal/agent"
	ctxbuilder "mindfs/server/internal/context"
	"mindfs/server/internal/session"
)

type ListSessionsInput struct {
	RootID string
}

type ListSessionsOutput struct {
	Sessions []*session.Session
}

func (s *Service) ListSessions(ctx context.Context, in ListSessionsInput) (ListSessionsOutput, error) {
	if err := s.ensureRegistry(); err != nil {
		return ListSessionsOutput{}, err
	}
	manager, err := s.Registry.GetSessionManager(in.RootID)
	if err != nil {
		return ListSessionsOutput{}, err
	}
	items, err := manager.List(ctx)
	if err != nil {
		return ListSessionsOutput{}, err
	}
	return ListSessionsOutput{Sessions: items}, nil
}

type CreateSessionInput struct {
	RootID string
	Input  session.CreateInput
}

func (s *Service) CreateSession(ctx context.Context, in CreateSessionInput) (*session.Session, error) {
	if err := s.ensureRegistry(); err != nil {
		return nil, err
	}
	manager, err := s.Registry.GetSessionManager(in.RootID)
	if err != nil {
		return nil, err
	}
	return manager.Create(ctx, in.Input)
}

type GetSessionInput struct {
	RootID string
	Key    string
}

func (s *Service) GetSession(ctx context.Context, in GetSessionInput) (*session.Session, error) {
	if err := s.ensureRegistry(); err != nil {
		return nil, err
	}
	manager, err := s.Registry.GetSessionManager(in.RootID)
	if err != nil {
		return nil, err
	}
	return manager.Get(ctx, in.Key)
}

type ResumeSessionInput struct {
	RootID string
	Key    string
}

func (s *Service) ResumeSession(ctx context.Context, in ResumeSessionInput) (*session.Session, error) {
	if err := s.ensureRegistry(); err != nil {
		return nil, err
	}
	manager, err := s.Registry.GetSessionManager(in.RootID)
	if err != nil {
		return nil, err
	}
	return manager.Resume(ctx, in.Key)
}

type CloseSessionInput struct {
	RootID string
	Key    string
}

func (s *Service) CloseSession(ctx context.Context, in CloseSessionInput) (*session.Session, error) {
	if err := s.ensureRegistry(); err != nil {
		return nil, err
	}
	manager, err := s.Registry.GetSessionManager(in.RootID)
	if err != nil {
		return nil, err
	}
	closed, err := manager.Close(ctx, in.Key)
	if err != nil {
		return nil, err
	}
	s.Registry.ReleaseFileWatcher(in.RootID, in.Key)
	return closed, nil
}

type AddExchangeInput struct {
	RootID  string
	Key     string
	Role    string
	Content string
}

func (s *Service) AddExchange(ctx context.Context, in AddExchangeInput) (*session.Session, error) {
	if err := s.ensureRegistry(); err != nil {
		return nil, err
	}
	manager, err := s.Registry.GetSessionManager(in.RootID)
	if err != nil {
		return nil, err
	}
	return manager.AddExchange(ctx, in.Key, in.Role, in.Content)
}

type BuildPromptInput struct {
	Session       *session.Session
	Manager       *session.Manager
	Message       string
	ClientContext ctxbuilder.ClientContext
	IsInitial     bool
}

func (s *Service) BuildPrompt(in BuildPromptInput) string {
	if !in.IsInitial {
		return ctxbuilder.BuildUserPrompt(in.Message, ctxbuilder.ClientContext{
			Selection: in.ClientContext.Selection,
		})
	}
	if in.Session == nil || in.Manager == nil {
		return ctxbuilder.BuildUserPrompt(in.Message, in.ClientContext)
	}
	serverCtx, err := ctxbuilder.BuildServerContext(
		in.Session.Type,
		in.Manager.Root(),
		in.ClientContext.CurrentView,
	)
	if err != nil {
		return ctxbuilder.BuildUserPrompt(in.Message, in.ClientContext)
	}
	serverPrompt := ctxbuilder.BuildServerPrompt(in.Session.Type, serverCtx)
	userPrompt := ctxbuilder.BuildUserPrompt(in.Message, in.ClientContext)
	if serverPrompt == "" {
		return userPrompt
	}
	return serverPrompt + "\n\n" + userPrompt
}

type AppendAgentReplyInput struct {
	RootID  string
	Key     string
	Content string
}

func (s *Service) AppendAgentReply(ctx context.Context, in AppendAgentReplyInput) error {
	if in.Content == "" {
		return nil
	}
	if err := s.ensureRegistry(); err != nil {
		return err
	}
	manager, err := s.Registry.GetSessionManager(in.RootID)
	if err != nil {
		return err
	}
	_, err = manager.AddExchange(ctx, in.Key, "agent", in.Content)
	return err
}

type SendMessageInput struct {
	RootID    string
	Key       string
	Content   string
	ClientCtx ctxbuilder.ClientContext
	OnUpdate  func(agent.Event)
}

func (s *Service) SendMessage(ctx context.Context, in SendMessageInput) error {
	start := time.Now()
	log.Printf("[session/send] begin root=%s session=%s content_chars=%d", in.RootID, in.Key, len(in.Content))
	if err := s.ensureRegistry(); err != nil {
		return err
	}
	t0 := time.Now()
	manager, err := s.Registry.GetSessionManager(in.RootID)
	if err != nil {
		return err
	}
	log.Printf("[session/send] get_manager session=%s duration_ms=%d", in.Key, time.Since(t0).Milliseconds())
	t1 := time.Now()
	before, err := manager.Get(ctx, in.Key)
	if err != nil {
		return err
	}
	log.Printf("[session/send] load_session_before session=%s duration_ms=%d", in.Key, time.Since(t1).Milliseconds())
	isInitial := len(before.Exchanges) == 0
	t2 := time.Now()
	_, err = manager.AddExchange(ctx, in.Key, "user", in.Content)
	if err != nil {
		return err
	}
	log.Printf("[session/send] append_user_exchange session=%s duration_ms=%d", in.Key, time.Since(t2).Milliseconds())
	t3 := time.Now()
	current, err := manager.Get(ctx, in.Key)
	if err != nil {
		return err
	}
	log.Printf("[session/send] load_session_after session=%s duration_ms=%d", in.Key, time.Since(t3).Milliseconds())
	agentPool := s.Registry.GetAgentPool()
	if agentPool == nil {
		return nil
	}
	t4 := time.Now()
	watcher, watcherErr := s.Registry.GetFileWatcher(in.RootID, manager)
	if watcherErr != nil {
		log.Printf("[watcher] root=%s session=%s get_failed err=%v", in.RootID, current.Key, watcherErr)
	}
	if watcher != nil {
		watcher.RegisterSession(current.Key)
		watcher.MarkSessionActive(current.Key)
	} else {
		log.Printf("[watcher] root=%s session=%s unavailable", in.RootID, current.Key)
	}
	log.Printf("[session/send] prepare_watcher session=%s duration_ms=%d", in.Key, time.Since(t4).Milliseconds())
	root := manager.Root()
	rootAbs, _ := root.RootDir()
	t5 := time.Now()
	sess, err := agentPool.GetOrCreate(ctx, current.Key, current.Agent, rootAbs)
	if err != nil {
		if prober := s.Registry.GetProber(); prober != nil {
			prober.ReportFailure(current.Agent, err)
		}
		return err
	}
	log.Printf("[session/send] get_or_create_agent_session session=%s agent=%s duration_ms=%d", in.Key, current.Agent, time.Since(t5).Milliseconds())

	t6 := time.Now()
	prompt := s.BuildPrompt(BuildPromptInput{
		Session:       current,
		Manager:       manager,
		Message:       in.Content,
		ClientContext: in.ClientCtx,
		IsInitial:     isInitial,
	})
	log.Printf("[session/send] build_prompt session=%s prompt_chars=%d duration_ms=%d", in.Key, len(prompt), time.Since(t6).Milliseconds())
	var responseText string
	sess.OnUpdate(func(update agent.Event) {
		if update.Type == agent.EventTypeToolCall {
			if toolCall, ok := update.Data.(agent.ToolCall); ok && toolCall.IsWriteOperation() {
				for _, path := range toolCall.GetAffectedPaths() {
					if watcher != nil {
						watcher.RecordPendingWrite(current.Key, path)
						watcher.RecordSessionFile(current.Key, path)
					}
				}
			}
		}
		if update.Type == agent.EventTypeMessageChunk {
			if chunk, ok := update.Data.(agent.MessageChunk); ok {
				responseText += chunk.Content
			}
		}
		if watcher != nil {
			watcher.MarkSessionActive(current.Key)
		}
		if in.OnUpdate != nil {
			in.OnUpdate(update)
		}
	})
	t7 := time.Now()
	if err := sess.SendMessage(ctx, prompt); err != nil {
		if prober := s.Registry.GetProber(); prober != nil {
			prober.ReportFailure(current.Agent, err)
		}
		return err
	}
	log.Printf("[session/send] agent_send_message_done session=%s duration_ms=%d", in.Key, time.Since(t7).Milliseconds())
	if prober := s.Registry.GetProber(); prober != nil {
		prober.ReportSuccess(current.Agent)
	}
	t8 := time.Now()
	err = s.AppendAgentReply(ctx, AppendAgentReplyInput{
		RootID:  in.RootID,
		Key:     in.Key,
		Content: responseText,
	})
	log.Printf("[session/send] append_agent_reply session=%s duration_ms=%d", in.Key, time.Since(t8).Milliseconds())
	if err != nil {
		return err
	}
	log.Printf("[session/send] done root=%s session=%s total_ms=%d", in.RootID, in.Key, time.Since(start).Milliseconds())
	return nil
}
